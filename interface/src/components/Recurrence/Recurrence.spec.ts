import { startOfMonth } from 'date-fns';

import Recurrence from './Recurrence';

import { RRule } from 'rrule';

describe('recurrence rules', () => {
  it('will compare equality regardless of dtstart', () => {
    // Different dtstarts, but same rule.
    const inputRule = 'DTSTART:20211231T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1';

    const recurrence = new Recurrence({
      rule: new RRule({
        dtstart: startOfMonth(new Date()),
        freq: RRule.MONTHLY,
        interval: 1,
        bymonthday: [15, -1],
      }),
    });

    expect(recurrence.equalRule(inputRule)).toBeTruthy();
  });

  it('will assert difference', () => {
    // Different rules now.
    const inputRule = 'DTSTART:20211231T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=3;BYMONTHDAY=1';

    const recurrence = new Recurrence({
      rule: new RRule({
        dtstart: startOfMonth(new Date()),
        freq: RRule.MONTHLY,
        interval: 1,
        bymonthday: [15, -1],
      }),
    });

    expect(recurrence.equalRule(inputRule)).toBeFalsy();
  });

  it('will return false for invalid rule', () => {
    const recurrence = new Recurrence({
      rule: new RRule({
        dtstart: startOfMonth(new Date()),
        freq: RRule.MONTHLY,
        interval: 1,
        bymonthday: [15, -1],
      }),
    });

    expect(recurrence.equalRule('bogus')).toBeFalsy();
  });
});
