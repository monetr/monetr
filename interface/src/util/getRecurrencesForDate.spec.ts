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

  it('will return an empty array when given a null input', () => {
    const result = getRecurrencesForDate(null, getTimezone());
    expect(result).toHaveLength(0);
  });

  it('will include 1st and 15th rule when the first day of the month is selected', () => {
    const timezone = tz(getTimezone());
    // First day of March 2026
    const input = timezone(parse('2026-03-01', 'yyyy-MM-dd', new Date()));

    const result = getRecurrencesForDate(input, getTimezone());

    const firstAndFifteenth = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1,15'),
    );
    if (!firstAndFifteenth) {
      console.log(JSON.stringify(result));
    }
    expect(firstAndFifteenth).not.toBeUndefined();

    // 15th and last should NOT appear when only the 1st is selected
    const fifteenthAndLast = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1'),
    );
    expect(fifteenthAndLast).toBeUndefined();
  });

  it('will not include combo rules for a regular mid-month day', () => {
    const timezone = tz(getTimezone());
    // March 10 is not the 1st, 15th, or last day of the month
    const input = timezone(parse('2026-03-10', 'yyyy-MM-dd', new Date()));

    const result = getRecurrencesForDate(input, getTimezone());

    // Only the 7 base rules should be returned
    expect(result).toHaveLength(7);

    const firstAndFifteenth = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1,15'),
    );
    expect(firstAndFifteenth).toBeUndefined();

    const fifteenthAndLast = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1'),
    );
    expect(fifteenthAndLast).toBeUndefined();
  });

  it('will return weekly rules for the correct day of the week', () => {
    const timezone = tz(getTimezone());
    // March 10, 2026 is a Tuesday
    const input = timezone(parse('2026-03-10', 'yyyy-MM-dd', new Date()));

    const result = getRecurrencesForDate(input, getTimezone());

    const everyTuesday = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=TU'),
    );
    if (!everyTuesday) {
      console.log(JSON.stringify(result));
    }
    expect(everyTuesday).not.toBeUndefined();

    const everyOtherTuesday = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=TU'),
    );
    expect(everyOtherTuesday).not.toBeUndefined();
  });

  it('will return every other month, quarterly, and every 6 month rules', () => {
    const timezone = tz(getTimezone());
    const input = timezone(parse('2026-03-10', 'yyyy-MM-dd', new Date()));

    const result = getRecurrencesForDate(input, getTimezone());

    const everyOtherMonth = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=MONTHLY;INTERVAL=2;BYMONTHDAY=10'),
    );
    expect(everyOtherMonth).not.toBeUndefined();

    const everyQuarter = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=MONTHLY;INTERVAL=3;BYMONTHDAY=10'),
    );
    expect(everyQuarter).not.toBeUndefined();

    const everySixMonths = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=MONTHLY;INTERVAL=6;BYMONTHDAY=10'),
    );
    expect(everySixMonths).not.toBeUndefined();
  });

  it('will use BYMONTHDAY=-1 for the last day of February', () => {
    const timezone = tz(getTimezone());
    // Last day of February 2026 (non-leap year)
    const input = timezone(parse('2026-02-28', 'yyyy-MM-dd', new Date()));

    const result = getRecurrencesForDate(input, getTimezone());

    const lastDayOfEveryMonth = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=-1'),
    );
    if (!lastDayOfEveryMonth) {
      console.log(JSON.stringify(result));
    }
    expect(lastDayOfEveryMonth).not.toBeUndefined();

    // Should not hardcode 28 since that is not always the last day of the month
    const lastDayWrong = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28'),
    );
    expect(lastDayWrong).toBeUndefined();

    // Last day of February also qualifies for the 15th and last rule
    const fifteenthAndLast = result.find(item =>
      item.ruleString().includes('RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1'),
    );
    expect(fifteenthAndLast).not.toBeUndefined();
  });
});
