import { tz } from '@date-fns/tz';
import { parse } from 'date-fns';

import getRecurrencesForDate from '@monetr/interface/util/getRecurrencesForDate';
import { getTimezone } from '@monetr/interface/util/locale';

describe('get recurrences for date', () => {
  it('will return the last day of every month when the last day is selected', () => {
    const timezone = tz(getTimezone());
    // Last day of march should generate a rule for the last day of every month
    const input = timezone(parse('2026-03-31', 'yyyy-MM-dd', new Date()));

    const result = getRecurrencesForDate(input, getTimezone());

    const lastDayOfEveryMonth = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=-1'),
    );
    if (!lastDayOfEveryMonth) {
      console.log(JSON.stringify(result));
    }
    // If this rule exists in the result then that means we are properly building the last day of every month rule.
    expect(lastDayOfEveryMonth).not.toBeUndefined();

    const lastDayOfEveryMonthWrong = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=31'),
    );
    // Make sure we don't see the old rule which was incorrect.
    expect(lastDayOfEveryMonthWrong).toBeUndefined();
  });

  it('will return first and fifthteenth and fifthteenth and last', () => {
    const timezone = tz(getTimezone());
    // Last day of march should generate a rule for the last day of every month
    const input = timezone(parse('2026-03-15', 'yyyy-MM-dd', new Date()));

    const result = getRecurrencesForDate(input, getTimezone());

    const firstAndFifthteenth = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1,15'),
    );
    const fifthteenthAndLast = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1'),
    );
    if (!fifthteenthAndLast) {
      console.log(JSON.stringify(result));
    }
    expect(firstAndFifthteenth).not.toBeUndefined();
    expect(fifthteenthAndLast).not.toBeUndefined();
  });
});
