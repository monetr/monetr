import { tz } from '@date-fns/tz';
import { endOfMonth, format, getDate, getMonth, isEqual, startOfDay, startOfMonth } from 'date-fns';
import { RRule, type Weekday } from 'rrule';

import Recurrence from '@monetr/interface/models/Recurrence';
import parseDate from '@monetr/interface/util/parseDate';

export default function getRecurrencesForDate(
  inputDate: Date | string | null,
  timezoneString: string,
): Array<Recurrence> {
  const date = parseDate(inputDate);
  if (!date) {
    return [];
  }

  const timezone = tz(timezoneString);
  const input = startOfDay(date, {
    in: timezone,
  });
  const endOfMonthDate = startOfDay(endOfMonth(input), {
    in: timezone,
  });
  const startOfMonthDate = startOfDay(startOfMonth(input), {
    in: timezone,
  });
  const isStartOfMonth = isEqual(input, startOfMonthDate);
  const isEndOfMonth = isEqual(input, endOfMonthDate);

  const weekdayString = format(input, 'EEEE');

  const ruleWeekday = getRuleDayOfWeek(input);

  const dayOfMonth = getDate(input);
  const dayStr = isEndOfMonth ? 'last day of the month' : ordinalSuffixOf(dayOfMonth);

  const rules = [
    new Recurrence({
      name: `Every ${weekdayString}`,
      rule: new RRule({
        dtstart: input,
        freq: RRule.WEEKLY,
        interval: 1,
        byweekday: [ruleWeekday],
      }),
    }),
    new Recurrence({
      name: `Every other ${weekdayString}`,
      rule: new RRule({
        dtstart: input,
        freq: RRule.WEEKLY,
        interval: 2,
        byweekday: [ruleWeekday],
      }),
    }),
    new Recurrence({
      name: `Every month on the ${dayStr}`,
      rule: new RRule({
        dtstart: input,
        freq: RRule.MONTHLY,
        interval: 1,
        bymonthday: [isEndOfMonth ? -1 : dayOfMonth],
      }),
    }),
    new Recurrence({
      name: `Every other month on the ${dayStr}`,
      rule: new RRule({
        dtstart: input,
        freq: RRule.MONTHLY,
        interval: 2,
        bymonthday: [isEndOfMonth ? -1 : dayOfMonth],
      }),
    }),
    new Recurrence({
      name: `Every 3 months (quarter) on the ${dayStr}`,
      rule: new RRule({
        dtstart: input,
        freq: RRule.MONTHLY,
        interval: 3,
        bymonthday: [isEndOfMonth ? -1 : dayOfMonth],
      }),
    }),
    new Recurrence({
      name: `Every 6 months on the ${dayStr}`,
      rule: new RRule({
        dtstart: input,
        freq: RRule.MONTHLY,
        interval: 6,
        bymonthday: [isEndOfMonth ? -1 : dayOfMonth],
      }),
    }),
    new Recurrence({
      name: `Every year on the ${ordinalSuffixOf(dayOfMonth)} of ${format(input, 'MMMM')}`,
      rule: new RRule({
        dtstart: input,
        freq: RRule.YEARLY,
        interval: 1,
        bymonth: [getMonth(input) + 1],
        bymonthday: [getDate(input)],
      }),
    }),
  ];

  if (isStartOfMonth || dayOfMonth === 15) {
    rules.push(
      new Recurrence({
        name: '1st and 15th of every month',
        rule: new RRule({
          dtstart: input,
          freq: RRule.MONTHLY,
          interval: 1,
          bymonthday: [1, 15],
        }),
      }),
    );
  }

  if (isEndOfMonth || dayOfMonth === 15) {
    rules.push(
      new Recurrence({
        name: '15th and last day of every month',
        rule: new RRule({
          dtstart: input,
          freq: RRule.MONTHLY,
          interval: 1,
          bymonthday: [15, -1],
        }),
      }),
    );
  }

  return rules;
}

function getRuleDayOfWeek(date: Date): Weekday {
  switch (format(date, 'EEEE')) {
    case 'Monday':
      return RRule.MO;
    case 'Tuesday':
      return RRule.TU;
    case 'Wednesday':
      return RRule.WE;
    case 'Thursday':
      return RRule.TH;
    case 'Friday':
      return RRule.FR;
    case 'Saturday':
      return RRule.SA;
    case 'Sunday':
      return RRule.SU;
    default:
      return RRule.SU;
  }
}

function ordinalSuffixOf(i: number) {
  const j = i % 10,
    k = i % 100;
  if (j === 1 && k !== 11) {
    return `${i}st`;
  }
  if (j === 2 && k !== 12) {
    return `${i}nd`;
  }
  if (j === 3 && k !== 13) {
    return `${i}rd`;
  }
  return `${i}th`;
}
