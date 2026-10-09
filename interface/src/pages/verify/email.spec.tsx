import { rs } from '@rstest/core';
import * as wouterActual from 'wouter' with { rstest: 'importActual' };

import { waitFor } from '@testing-library/react';

import VerifyEmail from '@monetr/interface/pages/verify/email';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';
import * as notifyActual from '@monetr/notify' with { rstest: 'importActual' };

const mockNavigate = rs.fn((_url: string) => {});
rs.mock('wouter', () => ({
  ...wouterActual,
  useLocation: () => ['/verify/email', mockNavigate],
}));

const mockEnqueueSnackbar = rs.fn();
rs.mock('@monetr/notify', () => ({
  ...notifyActual,
  useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

describe('verify email page', () => {
  let mockFetch: FetchMock;

  beforeEach(() => {
    mockFetch = new FetchMock();
    mockNavigate.mockReset();
    mockEnqueueSnackbar.mockReset();
  });
  afterEach(() => {
    mockFetch.reset();
  });
  afterAll(() => {
    mockFetch.restore();
  });

  it('will verify the email', async () => {
    mockFetch.onPost('/api/authentication/verify').reply(200, {
      nextUrl: '/login',
      message: 'Your email is now verified. Please login.',
    });

    testRenderer(<VerifyEmail />, { initialRoute: '/verify/email?token=01hym36e8ewaq0hxssb1m3k4ha' });

    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith('/login'));
    expect(mockFetch.history.post).toHaveLength(1);
    expect(mockFetch.history.post?.[0]).toMatchObject({
      url: '/api/authentication/verify',
      data: {
        token: '01hym36e8ewaq0hxssb1m3k4ha',
      },
    });
    expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Your email is now verified. Please login.', {
      variant: 'success',
      disableWindowBlurListener: true,
    });
  });

  it('will show an error when the token is not valid', async () => {
    mockFetch.onPost('/api/authentication/verify').reply(400, {
      error: 'Invalid email verification',
    });

    testRenderer(<VerifyEmail />, { initialRoute: '/verify/email?token=01hym36e8ewaq0hxssb1m3k4ha' });

    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith('/login'));
    expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Invalid email verification', {
      variant: 'error',
      disableWindowBlurListener: true,
    });
  });

  it('will show an error when there is no token', async () => {
    testRenderer(<VerifyEmail />, { initialRoute: '/verify/email' });

    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith('/login'));
    expect(mockFetch.history.post).toHaveLength(0);
    expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Email verification link is not valid.', {
      variant: 'error',
      disableWindowBlurListener: true,
    });
  });
});
