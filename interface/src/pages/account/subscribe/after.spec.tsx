import { rs } from '@rstest/core';
import * as wouterActual from 'wouter' with { rstest: 'importActual' };

import { waitFor } from '@testing-library/react';

import AfterCheckoutPage from '@monetr/interface/pages/account/subscribe/after';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';
import * as notifyActual from '@monetr/notify' with { rstest: 'importActual' };

const mockNavigate = rs.fn((_url: string) => {});
rs.mock('wouter', () => ({
  ...wouterActual,
  useLocation: () => ['/account/subscribe/after', mockNavigate],
}));

const mockEnqueueSnackbar = rs.fn();
rs.mock('@monetr/notify', () => ({
  ...notifyActual,
  useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

describe('after checkout page', () => {
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

  it('will redirect when the subscription is active', async () => {
    mockFetch.onGet('/api/billing/checkout/cs_test_a1b2c3d4e5').reply(200, {
      message: null,
      nextUrl: '/',
      isActive: true,
    });

    testRenderer(<AfterCheckoutPage />, { initialRoute: '/account/subscribe/after?session=cs_test_a1b2c3d4e5' });

    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith('/'));
    expect(mockEnqueueSnackbar).not.toHaveBeenCalled();
  });

  it('will show the message when the subscription is not active', async () => {
    mockFetch.onGet('/api/billing/checkout/cs_test_a1b2c3d4e5').reply(200, {
      message: 'Subscription is not active.',
      nextUrl: '/account/subscribe',
      isActive: false,
    });

    testRenderer(<AfterCheckoutPage />, { initialRoute: '/account/subscribe/after?session=cs_test_a1b2c3d4e5' });

    await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(1));
    expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Subscription is not active.', {
      variant: 'error',
      disableWindowBlurListener: true,
    });
    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it('will show an error when the checkout cannot be retrieved', async () => {
    mockFetch.onGet('/api/billing/checkout/cs_test_a1b2c3d4e5').reply(500, {
      error: 'Failed to retrieve checkout session',
    });

    testRenderer(<AfterCheckoutPage />, { initialRoute: '/account/subscribe/after?session=cs_test_a1b2c3d4e5' });

    await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(1));
    expect(mockEnqueueSnackbar).toHaveBeenCalledWith(
      'Unable to determine your subscription state, please contact support@monetr.app',
      {
        variant: 'error',
        disableWindowBlurListener: true,
      },
    );
    expect(mockNavigate).not.toHaveBeenCalled();
  });
});
