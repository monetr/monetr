import { rs } from '@rstest/core';
import * as wouterActual from 'wouter' with { rstest: 'importActual' };
import { Route } from 'wouter';

import { waitFor } from '@testing-library/react';

import SimpleFinSetupAccounts from '@monetr/interface/components/setup/simplefin/SimpleFinSetupAccounts';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';
import * as notifyActual from '@monetr/notify' with { rstest: 'importActual' };

const mockNavigate = rs.fn((_url: string) => {});
rs.mock('wouter', () => ({
  ...wouterActual,
  useLocation: () => ['/setup/simplefin/sfx_01hy4rbb1gjdek7h2xmgy5pnwk', mockNavigate],
}));

const mockEnqueueSnackbar = rs.fn();
rs.mock('@monetr/notify', () => ({
  ...notifyActual,
  useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

describe('simplefin setup accounts', () => {
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

  it('will redirect when the link is not pending', async () => {
    mockFetch.onPost('/api/simplefin/link/sfx_01hy4rbb1gjdek7h2xmgy5pnwk/bank_accounts/refresh').reply(204);
    mockFetch.onGet('/api/simplefin/link/sfx_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      simpleFinLinkId: 'sfx_01hy4rbb1gjdek7h2xmgy5pnwk',
      name: 'SimpleFIN',
      apiUrl: 'https://beta-bridge.simplefin.org/simplefin',
      status: 'active',
      lastErrors: [],
      lastManualSync: null,
      lastSuccessfulUpdate: '2026-03-16T06:00:00Z',
      lastAttemptedUpdate: '2026-03-16T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
      createdAt: '2026-03-15T06:00:00Z',
      deletedAt: null,
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
    });
    mockFetch.onGet('/api/simplefin/link/sfx_01hy4rbb1gjdek7h2xmgy5pnwk/bank_accounts').reply(200, []);

    testRenderer(<Route component={SimpleFinSetupAccounts} path='/setup/simplefin/:simpleFinLinkId' />, {
      initialRoute: '/setup/simplefin/sfx_01hy4rbb1gjdek7h2xmgy5pnwk',
    });

    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith('/'));
    expect(mockEnqueueSnackbar).toHaveBeenCalledWith(
      'SimpleFIN link is not in a pending status and cannot be setup this way. Redirecting you...',
      {
        variant: 'warning',
        disableWindowBlurListener: true,
      },
    );
  });
});
