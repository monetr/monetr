import React from 'react';
import { waitFor } from '@testing-library/react';
import { rest } from 'msw';

import FundingDetails from './details';
import testRenderer from '@monetr/interface/testutils/renderer';
import { server } from '@monetr/interface/testutils/server';

describe('funding schedule details view', () => {
  it('will render with adjusted weekend', async () => {
    server.use(
      rest.get('/api/bank_accounts/12', (_req, res, ctx) => {
        return res(ctx.json({
          'bankAccountId': 12,
          'linkId': 4,
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
      rest.get('/api/bank_accounts/12/funding_schedules/1', (_req, res, ctx) => {
        return res(ctx.json({
          'bankAccountId': 12,
          'dateStarted': '2023-02-28T06:00:00Z',
          'description': '15th and last day of every month',
          'estimatedDeposit': null,
          'excludeWeekends': true,
          'fundingScheduleId': 1,
          'lastOccurrence': '2023-09-29T05:00:00Z',
          'name': 'Elliot\'s Contribution',
          'nextOccurrence': '2023-10-13T05:00:00Z',
          'nextOccurrenceOriginal': '2023-10-15T05:00:00Z',
          'ruleset': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
          'waitForDeposit': false,
        }));
      }),
    );

    const world = testRenderer(<FundingDetails />, { initialRoute: '/bank/12/funding/1/details' });

    await waitFor(() => expect(world.getByTestId('funding-details-date-picker')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('funding-schedule-weekend-notice')).toBeVisible());
  });

  it('will render without adjusted weekend', async () => {
    server.use(
      rest.get('/api/bank_accounts/12', (_req, res, ctx) => {
        return res(ctx.json({
          'bankAccountId': 12,
          'linkId': 4,
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
      rest.get('/api/bank_accounts/12/funding_schedules/1', (_req, res, ctx) => {
        return res(ctx.json({
          'bankAccountId': 12,
          'dateStarted': '2023-02-28T06:00:00Z',
          'description': '15th and last day of every month',
          'estimatedDeposit': null,
          'excludeWeekends': false,
          'fundingScheduleId': 1,
          'lastOccurrence': '2023-09-30T05:00:00Z',
          'name': 'Elliot\'s Contribution',
          'nextOccurrence': '2023-10-15T05:00:00Z',
          'nextOccurrenceOriginal': '2023-10-15T05:00:00Z',
          'ruleset': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
          'waitForDeposit': false,
        }));
      }),
    );

    const world = testRenderer(<FundingDetails />, { initialRoute: '/bank/12/funding/1/details' });

    await waitFor(() => expect(world.getByTestId('funding-details-date-picker')).toBeVisible());
    await waitFor(() => expect(world.queryByTestId('funding-schedule-weekend-notice')).not.toBeInTheDocument());
  });
});
