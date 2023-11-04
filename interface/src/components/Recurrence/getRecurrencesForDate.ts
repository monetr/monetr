/* eslint-disable id-length */
import { endOfMonth, format, getDate, getMonth, isEqual, parseJSON, startOfDay, startOfMonth } from 'date-fns';

import Recurrence from '@monetr/interface/components/Recurrence/Recurrence';

import { RRule, Weekday } from 'rrule';

export default function getRecurrencesForDate(inputDate: Date | string | null): Array<Recurrence> {
  let date: Date;
  if (typeof inputDate === 'string') {
    date = parseJSON(inputDate);
  } else if (!inputDate) {
    return [];
  } else {
    date = inputDate;
  }

  const input = startOfDay(date);
  const endOfMonthDate = startOfDay(endOfMonth(input));
  const startOfMonthDate = startOfDay(startOfMonth(input));
  const isStartOfMonth = isEqual(input, startOfMonthDate);
  const isEndOfMonth = isEqual(input, endOfMonthDate);

  const weekdayString = format(input, 'EEEE');

  const ruleWeekday = getRuleDayOfWeek(input);

  const dayStr = isEndOfMonth ? ' last day of the month' : ordinalSuffixOf(getDate(input));

  const rules = [
    new Recurrence({
      name: `Every ${ weekdayString }`,
      rule: new RRule({
        dtstart: input,
        freq: RRule.WEEKLY,
        interval: 1,
        byweekday: [ruleWeekday],
      }),
    }),
    new Recurrence({
      name: `Every other ${ weekdayString }`,
      rule: new RRule({
        dtstart: input,
        freq: RRule.WEEKLY,
        interval: 2,
        byweekday: [ruleWeekday],
      }),
    }),
    new Recurrence({
      name: `Every month on the ${ dayStr }`,
      rule: new RRule({
        dtstart: input,
        freq: RRule.MONTHLY,
        interval: 1,
        bymonthday: getDate(input),
      }),
    }),
    new Recurrence({
      name: `Every other month on the ${ dayStr }`,
      rule: new RRule({
        dtstart: input,
        freq: RRule.MONTHLY,
        interval: 2,
        bymonthday: getDate(input),
      }),
    }),
    new Recurrence({
      name: `Every 3 months (quarter) on the ${ dayStr }`,
      rule: new RRule({
        dtstart: input,
        freq: RRule.MONTHLY,
        interval: 3,
        bymonthday: getDate(input),
      }),
    }),
    new Recurrence({
      name: `Every 6 months on the ${ dayStr }`,
      rule: new RRule({
        dtstart: input,
        freq: RRule.MONTHLY,
        interval: 6,
        bymonthday: getDate(input),
      }),
    }),
    new Recurrence({
      name: `Every year on the ${ ordinalSuffixOf(getDate(input)) } of ${ format(input, 'MMMM') }`,
      rule: new RRule({
        dtstart: input,
        freq: RRule.YEARLY,
        interval: 1,
        bymonth: getMonth(input) + 1,
        bymonthday: getDate(input),
      }),
    }),
  ];

  if (isStartOfMonth || getDate(input) === 15) {
    rules.push(new Recurrence({
      name: '1st and 15th of every month',
      rule: new RRule({
        dtstart: input,
        freq: RRule.MONTHLY,
        interval: 1,
        bymonthday: [1, 15],
      }),
    }));
  }

  if (isEndOfMonth || getDate(input) === 15) {
    rules.push(new Recurrence({
      name: '15th and last day of every month',
      rule: new RRule({
        dtstart: input,
        freq: RRule.MONTHLY,
        interval: 1,
        bymonthday: [15, -1],
      }),
    }));
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
  const j = i % 10, k = i % 100;
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
