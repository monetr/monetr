import { act } from '@testing-library/react-hooks';
import { parseJSON } from 'date-fns';
import { rest } from 'msw';

import { FundingScheduleUpdateResponse, useCreateFundingSchedule, useFundingSchedule, useFundingSchedulesSink, useUpdateFundingSchedule } from '@monetr/interface/hooks/fundingSchedules';
import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import testRenderHook from '@monetr/interface/testutils/hooks';
import { server } from '@monetr/interface/testutils/server';

describe('funding schedule hooks', () => {
  describe('read funding schedules', () => {
    it('will request all funding schedules', async () => {
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
        rest.get('/api/bank_accounts/12/funding_schedules', (_req, res, ctx) => {
          return res(ctx.json([
            {
              'bankAccountId': 12,
              'dateStarted': '2023-02-28T06:00:00Z',
              'description': '15th and last day of every month',
              'estimatedDeposit': null,
              'excludeWeekends': true,
              'fundingScheduleId': 3,
              'lastOccurrence': '2023-07-14T05:00:00Z',
              'name': 'Elliot\'s Contribution',
              'nextOccurrence': '2023-07-31T05:00:00Z',
              'nextOccurrenceOriginal': '2023-07-31T05:00:00Z',
              'rule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
              'waitForDeposit': false,
            },
          ]));
        }),
      );

      const world = testRenderHook(useFundingSchedulesSink, { initialRoute: '/bank/12/funding' });
      await world.waitFor(() => expect(world.result.current.isLoading).toBeTruthy());
      await world.waitForNextUpdate();
      await world.waitFor(() => expect(world.result.current.isFetching).toBeTruthy());
      await world.waitForNextUpdate();
      await world.waitFor(() => expect(world.result.current.data).toBeDefined());
      await world.waitFor(() => expect(world.result.current.data).toHaveLength(1));
    });

    it('will request a single funding schedule', async () => {
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
        rest.get('/api/bank_accounts/12/funding_schedules/3', (_req, res, ctx) => {
          return res(ctx.json({
            'bankAccountId': 12,
            'dateStarted': '2023-02-28T06:00:00Z',
            'description': '15th and last day of every month',
            'estimatedDeposit': null,
            'excludeWeekends': true,
            'fundingScheduleId': 3,
            'lastOccurrence': '2023-07-14T05:00:00Z',
            'name': 'Elliot\'s Contribution',
            'nextOccurrence': '2023-07-31T05:00:00Z',
            'nextOccurrenceOriginal': '2023-07-31T05:00:00Z',
            'rule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
            'waitForDeposit': false,
          }));
        }),
      );

      const world = testRenderHook(() => useFundingSchedule(3), { initialRoute: '/bank/12/funding' });
      await world.waitFor(() => expect(world.result.current.isLoading).toBeTruthy());
      await world.waitForNextUpdate();
      await world.waitFor(() => expect(world.result.current.isFetching).toBeTruthy());
      await world.waitForNextUpdate();
      await world.waitFor(() => expect(world.result.current.data).toBeDefined());
      await world.waitFor(() => expect(world.result.current.data?.fundingScheduleId).toBe(3));
    });
  });

  describe('create funding schedule', () => {
    it('will create a funding schedule', async () => {
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
        rest.post('/api/bank_accounts/12/funding_schedules', (_req, res, ctx) => {
          return res(ctx.json({
            'bankAccountId': 12,
            'dateStarted': '2023-02-28T06:00:00Z',
            'description': '15th and last day of every month',
            'estimatedDeposit': null,
            'excludeWeekends': true,
            'fundingScheduleId': 3,
            'lastOccurrence': '2023-07-14T05:00:00Z',
            'name': 'Elliot\'s Contribution',
            'nextOccurrence': '2023-07-31T05:00:00Z',
            'nextOccurrenceOriginal': '2023-07-31T05:00:00Z',
            'rule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
            'waitForDeposit': false,
          }));
        }),
      );

      const world = testRenderHook(useCreateFundingSchedule, { initialRoute: '/bank/12/funding' });
      let result: FundingSchedule;
      await act(async () => {
        result = await world.result.current(new FundingSchedule({
          bankAccountId: 12,
          description: 'something',
          name: 'Elliot\'s Contribution',
          nextOccurrence: parseJSON('2023-07-31T05:00:00Z'),
          ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
          estimatedDeposit: null,
          excludeWeekends: true,
        }));
      });
      expect(result).toBeDefined();
      expect(result.fundingScheduleId).toBe(3);
    });

    it('it will fail to create a funding schedule', async () => {
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
        rest.post('/api/bank_accounts/12/funding_schedules', (_req, res, ctx) => {
          return res(ctx.status(400), ctx.json({
            'error': 'Invalid funding schedule or something',
          }));
        }),
      );

      const world = testRenderHook(useCreateFundingSchedule, { initialRoute: '/bank/12/funding' });
      await act(async () => {
        expect(async () => {
          return world.result.current(new FundingSchedule({
            bankAccountId: 12,
            description: 'something',
            name: 'Elliot\'s Contribution',
            nextOccurrence: parseJSON('2023-07-31T05:00:00Z'),
            ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
            estimatedDeposit: null,
            excludeWeekends: true,
          }));
        }).rejects.toMatchObject({
          message: 'Request failed with status code 400',
        });
      });
    });
  });

  describe('update funding schedule', () => {
    it('will update a funding schedule', async () => {
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
        rest.put('/api/bank_accounts/12/funding_schedules/3', (_req, res, ctx) => {
          return res(ctx.json({
            'fundingSchedule': {
              'bankAccountId': 12,
              'dateStarted': '2023-02-28T06:00:00Z',
              'description': '15th and last day of every month',
              'estimatedDeposit': null,
              'excludeWeekends': true,
              'fundingScheduleId': 3,
              'lastOccurrence': '2023-07-14T05:00:00Z',
              'name': 'Elliot\'s Contribution',
              'nextOccurrence': '2023-07-31T05:00:00Z',
              'nextOccurrenceOriginal': '2023-07-31T05:00:00Z',
              'rule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
              'waitForDeposit': false,
            },
            'spending': [],
          }));
        }),
      );

      const world = testRenderHook(useUpdateFundingSchedule, { initialRoute: '/bank/12/funding' });
      let result: FundingScheduleUpdateResponse;
      await act(async () => {
        result = await world.result.current(new FundingSchedule({
          fundingScheduleId: 3,
          bankAccountId: 12,
          description: 'something',
          name: 'Elliot\'s Contribution',
          nextOccurrence: parseJSON('2023-07-31T05:00:00Z'),
          ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
          estimatedDeposit: null,
          excludeWeekends: true,
        }));
      });
      expect(result).toBeDefined();
      expect(result.fundingSchedule.fundingScheduleId).toBe(3);
      expect(result.spending.length).toBe(0);
    });

    it('it will fail to update a funding schedule', async () => {
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
        rest.put('/api/bank_accounts/12/funding_schedules/3', (_req, res, ctx) => {
          return res(ctx.status(400), ctx.json({
            'error': 'Invalid funding schedule or something',
          }));
        }),
      );

      const world = testRenderHook(useUpdateFundingSchedule, { initialRoute: '/bank/12/funding' });
      await act(async () => {
        await expect(async () => {
          return world.result.current(new FundingSchedule({
            fundingScheduleId: 3,
            bankAccountId: 12,
            description: 'something',
            name: 'Elliot\'s Contribution',
            nextOccurrence: parseJSON('2023-07-31T05:00:00Z'),
            ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
            estimatedDeposit: null,
            excludeWeekends: true,
          }));
        }).rejects.toMatchObject({
          message: 'Request failed with status code 400',
        });
      });
    });
  });
});

