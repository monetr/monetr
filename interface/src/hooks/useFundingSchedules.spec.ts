import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import { useFundingSchedules } from '@monetr/interface/hooks/useFundingSchedules';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('will list all funding schedules', () => {
  let mockAxios: MockAdapter;

  beforeEach(() => {
    mockAxios = new MockAdapter(monetrClient);
  });
  afterEach(() => {
    mockAxios.reset();
  });

  afterAll(() => mockAxios.restore());

  it('will request all funding schedules', async () => {
    mockAxios.onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx').reply(200, {
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
    });

    mockAxios.onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/funding_schedules').reply(200, [
      {
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
      },
    ]);

    const world = testRenderHook(useFundingSchedules, {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/funding',
    });
    await world.waitFor(() => expect(world.result.current.isLoading).toBeTruthy());
    // await world.waitForNextUpdate();
    await world.waitFor(() => expect(world.result.current.isFetching).toBeTruthy());
    // await world.waitForNextUpdate();
    await world.waitFor(() => expect(world.result.current.data).toBeDefined());
    await world.waitFor(() => expect(world.result.current.data).toHaveLength(1));
  });
});
