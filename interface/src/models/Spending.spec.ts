
import Spending from '@monetr/interface/models/Spending';

import { describe, expect, it } from 'bun:test';
import { rrulestr } from 'rrule';

describe('spending', () => {
  it('will handle a spending object that crosses DST', () => {
    const object = {
      'spendingId': 'spnd_01fkvmp0y8cyrb8j37r040jbdg',
      'bankAccountId': 'bac_01gds1cs1pt848f73dxrdt20at',
      'fundingScheduleId': 'fund_01hym37k3krbj9s0cphqnc92s0',
      'spendingType': 0,
      'name': 'Quarterly Expense',
      'description': 'Every 3 months (quarter) on the 1st',
      'targetAmount': 20000,
      'currentAmount': 0,
      'usedAmount': 0,
      'ruleset': 'DTSTART:20230401T000000Z\nRRULE:FREQ=MONTHLY;INTERVAL=3;BYMONTHDAY=1',
      'lastSpentFrom': null,
      'lastRecurrence': '2024-10-01T05:00:00Z',
      'nextRecurrence': '2025-01-01T06:00:00Z',
      'nextContributionAmount': 2974,
      'isBehind': false,
      'isPaused': false,
      'createdAt': '2021-11-06T22:07:41Z',
    };
    const spending = new Spending(object as any);
    const rule = rrulestr(spending.ruleset, { tzid: 'America/Chicago' });
    const now = new Date('2024-10-08T22:15:04.541Z');
    const nextAfter = rule.after(now);
    expect(spending.nextRecurrence.toISOString()).toEqual(nextAfter.toISOString());
  });
});
