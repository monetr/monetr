import { tz } from '@date-fns/tz';
import { parse } from 'date-fns';

import getRecurrencesForDate from '@monetr/interface/util/getRecurrencesForDate';

describe('get recurrences for date', () => {
  it.skip('will return the last day of every month when the last day is selected', () => {
    const timezone = tz('America/Chicago');
    // Last day of march should generate a rule for the last day of every month
    const input = timezone(parse('2026-03-31', 'yyyy-MM-dd', new Date()));

    const result = getRecurrencesForDate(input, 'America/Chicago');

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
});
