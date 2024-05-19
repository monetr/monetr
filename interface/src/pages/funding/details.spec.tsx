import React from 'react';
import { waitFor } from '@testing-library/react';
import { rest } from 'msw';

import FundingDetails from './details';
import testRenderer from '@monetr/interface/testutils/renderer';
import { server } from '@monetr/interface/testutils/server';

describe('funding schedule details view', () => {
  it('will render with adjusted weekend', async () => {
    server.use(
      rest.get('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx', (_req, res, ctx) => {
        return res(ctx.json({
          'bankAccountId': 'bac_01hy4rcmadc01d2kzv7vynbxxx', // 12,
          'linkId': 'link_01hy4rbb1gjdek7h2xmgy5pnwk', // 4
          'availableBalance': 48635,
          'currentBalance': 48635,
          'mask': '2982',
          'name': 'Mercury Checking',
          'originalName': 'Mercury Checking',
          'officialName': 'Mercury Checking',
          'accountType': 'depository',
          'accountSubType': 'checking',
          'status': 'active',
          'lastUpdated': '2023-07-02T04:22:52.48118Z',
        }));
      }),
      rest.get('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/funding_schedules/fund_01hy4re7c1xc2v44cf6kx302jx', (_req, res, ctx) => {
        return res(ctx.json({
          'bankAccountId': 'bac_01hy4rcmadc01d2kzv7vynbxxx', // 12,
          'dateStarted': '2023-02-28T06:00:00Z',
          'description': '15th and last day of every month',
          'estimatedDeposit': null,
          'excludeWeekends': true,
          'fundingScheduleId': 'fund_01hy4re7c1xc2v44cf6kx302jx', // 3,
          'lastRecurrence': '2023-09-29T05:00:00Z',
          'name': 'Elliot\'s Contribution',
          'nextRecurrence': '2023-10-13T05:00:00Z',
          'nextRecurrenceOriginal': '2023-10-15T05:00:00Z',
          'ruleset': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
          'waitForDeposit': false,
        }));
      }),
    );

    const world = testRenderer(<FundingDetails />, { 
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/funding/fund_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    await waitFor(() => expect(world.getByTestId('funding-details-date-picker')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('funding-schedule-weekend-notice')).toBeVisible());
  });

  it('will render without adjusted weekend', async () => {
    server.use(
      rest.get('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx', (_req, res, ctx) => {
        return res(ctx.json({
          'bankAccountId': 'bac_01hy4rcmadc01d2kzv7vynbxxx', // 12,
          'linkId': 'link_01hy4rbb1gjdek7h2xmgy5pnwk', // 4
          'availableBalance': 48635,
          'currentBalance': 48635,
          'mask': '2982',
          'name': 'Mercury Checking',
          'originalName': 'Mercury Checking',
          'officialName': 'Mercury Checking',
          'accountType': 'depository',
          'accountSubType': 'checking',
          'status': 'active',
          'lastUpdated': '2023-07-02T04:22:52.48118Z',
        }));
      }),
      rest.get('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/funding_schedules/fund_01hy4re7c1xc2v44cf6kx302jx', (_req, res, ctx) => {
        return res(ctx.json({
          'bankAccountId': 'bac_01hy4rcmadc01d2kzv7vynbxxx', // 12,
          'linkId': 'link_01hy4rbb1gjdek7h2xmgy5pnwk', // 4
          'dateStarted': '2023-02-28T06:00:00Z',
          'description': '15th and last day of every month',
          'estimatedDeposit': null,
          'excludeWeekends': false,
          'fundingScheduleId': 'fund_01hy4re7c1xc2v44cf6kx302jx', // 3,
          'lastRecurrence': '2023-09-30T05:00:00Z',
          'name': 'Elliot\'s Contribution',
          'nextRecurrence': '2023-10-15T05:00:00Z',
          'nextRecurrenceOriginal': '2023-10-15T05:00:00Z',
          'ruleset': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
          'waitForDeposit': false,
        }));
      }),
    );

    const world = testRenderer(<FundingDetails />, { 
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/funding/fund_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    await waitFor(() => expect(world.getByTestId('funding-details-date-picker')).toBeVisible());
    await waitFor(() => expect(world.queryByTestId('funding-schedule-weekend-notice')).not.toBeInTheDocument());
  });
});
