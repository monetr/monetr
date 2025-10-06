import { act } from 'react';
import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import { PatchFundingScheduleResponse, usePatchFundingSchedule } from '@monetr/interface/hooks/usePatchFundingSchedule';
import testRenderHook from '@monetr/interface/testutils/hooks';
import parseDate from '@monetr/interface/util/parseDate';

describe('patch funding schedule', () => {
  let mockAxios: MockAdapter;

  beforeEach(() => {
    mockAxios = new MockAdapter(monetrClient);
  });
  afterEach(() => {
    mockAxios.reset();
  });

  afterAll(() => mockAxios.restore());

  it('will update a funding schedule', async () => {
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
    mockAxios.onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/funding_schedules/fund_01hy4re7c1xc2v44cf6kx302jx').reply(200, {
      'fundingSchedule': {
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
      'spending': [],
    });

    const world = testRenderHook(usePatchFundingSchedule, {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/funding',
    });
    let result: PatchFundingScheduleResponse;
    await act(async () => {
      result = await world.result.current({
        fundingScheduleId: 'fund_01hy4re7c1xc2v44cf6kx302jx', // 3,
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
        description: 'something',
        name: 'Elliot\'s Contribution',
        nextRecurrence: parseDate('2023-07-31T05:00:00Z'),
        ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
        estimatedDeposit: null,
        excludeWeekends: true,
      });
    });
    expect(result).toBeDefined();
    expect(result.fundingSchedule.fundingScheduleId).toBe('fund_01hy4re7c1xc2v44cf6kx302jx');
    expect(result.spending.length).toBe(0);
  });

  it('it will fail to update a funding schedule', async () => {
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
    mockAxios.onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/funding_schedules/fund_01hy4re7c1xc2v44cf6kx302jx').reply(400, {
      'error': 'Invalid request',
    });

    const world = testRenderHook(usePatchFundingSchedule, {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/funding',
    });
    await act(async () => {
      expect(world.result.current({
        fundingScheduleId: 'fund_01hy4re7c1xc2v44cf6kx302jx', // 3,
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
        description: 'something',
        name: 'Elliot\'s Contribution',
        nextRecurrence: parseDate('2023-07-31T05:00:00Z'),
        ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
        estimatedDeposit: null,
        excludeWeekends: true,
      })).rejects.toMatchObject({
        message: 'Request failed with status code 400',
        response: {
          data: {
            'error': 'Invalid request',
          },
        },
      });
    });
  });
});
