import { useFundingSchedule, useFundingSchedulesSink } from 'hooks/fundingSchedules';
import { rest } from 'msw';
import testRenderHook from 'testutils/hooks';
import { server } from 'testutils/server';

describe('funding schedule hooks', () => {
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

