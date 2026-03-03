import { parse } from 'date-fns';

import getRecurrencesForDate from '@monetr/interface/util/getRecurrencesForDate';

describe('get recurrences for date', () => {
  it('will return the last day of every month when the last day is selected', () => {
    // Last day of march should generate a rule for the last day of every month
    const input = parse('2026-03-31', 'yyyy-MM-dd', new Date());

    const result = getRecurrencesForDate(input, 'America/Chicago');

    const lastDayOfEveryMonth = result.find(
      item =>
        item.ruleString() ===
        `DTSTART:20260331T050000Z
RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=-1`,
    );
    // If this rule exists in the result then that means we are properly building the last day of every month rule.
    expect(lastDayOfEveryMonth).not.toBeUndefined();

    const lastDayOfEveryMonthWrong = result.find(
      item =>
        item.ruleString() ===
        `DTSTART:20260331T050000Z
RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=31`,
    );
    // Make sure we don't see the old rule which was incorrect.
    expect(lastDayOfEveryMonthWrong).toBeUndefined();
  });
});
