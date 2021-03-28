import Recurrence from "components/Recurrence/Recurrence";
import moment from "moment";
import RRule, { Weekday } from "rrule";
import { List } from 'immutable';

export default function getRecurrencesForDate(date: moment.Moment): List<Recurrence> {
  const input = date.clone().startOf('day');
  const endOfMonth = input.clone().endOf('month').startOf('day');
  const startOfMonth = input.clone().startOf('month').startOf('day');
  const isStartOfMonth = input.unix() === startOfMonth.unix();
  const isEndOfMonth = input.unix() === endOfMonth.unix();

  const weekdayString = input.format('dddd');

  const ruleWeekday = getRuleDayOfWeek(input);

  const dayStr = isEndOfMonth ? ' last day of the month' : ordinalSuffixOf(input.date());

  let rules = [
    new Recurrence({
      name: `Every ${ weekdayString }`,
      rule: new RRule({
        freq: RRule.WEEKLY,
        interval: 1,
        byweekday: [ruleWeekday],
      }),
    }),
    new Recurrence({
      name: `Every other ${ weekdayString }`,
      rule: new RRule({
        freq: RRule.WEEKLY,
        interval: 2,
        byweekday: [ruleWeekday],
      }),
    }),
    new Recurrence({
      name: `Every month on the ${ dayStr }`,
      rule: new RRule({
        freq: RRule.MONTHLY,
        interval: 1,
        bymonthday: input.date(),
      }),
    }),
    new Recurrence({
      name: `Every other month on the ${ dayStr }`,
      rule: new RRule({
        freq: RRule.MONTHLY,
        interval: 2,
        bymonthday: input.date(),
      }),
    }),
    new Recurrence({
      name: `Every 3 months (quarter) on the ${ dayStr }`,
      rule: new RRule({
        freq: RRule.MONTHLY,
        interval: 3,
        bymonthday: input.date(),
      }),
    }),
    new Recurrence({
      name: `Every 6 months on the ${ dayStr }`,
      rule: new RRule({
        freq: RRule.MONTHLY,
        interval: 3,
        bymonthday: input.date(),
      }),
    }),
    new Recurrence({
      name: `Every year on the ${ ordinalSuffixOf(input.date()) } of ${ input.format('MMMM') }`,
      rule: new RRule({
        freq: RRule.YEARLY,
        interval: 3,
        bymonthday: input.date(),
      }),
    }),
  ];

  if (isStartOfMonth) {
    rules.push(new Recurrence({
      name: `On the 1st and 15th of every month`,
      rule: new RRule({
        freq: RRule.MONTHLY,
        interval: 1,
        bymonthday: [1, 15],
      })
    }));
  }

  if (isEndOfMonth) {
    rules.push(new Recurrence({
      name: `On the 15th and last day of every month`,
      rule: new RRule({
        freq: RRule.MONTHLY,
        interval: 1,
        bymonthday: [15, -1],
      })
    }));
  }

  return List<Recurrence>(rules);
}

function getRuleDayOfWeek(date: moment.Moment): Weekday {
  switch (date.format('dddd')) {
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

function ordinalSuffixOf(i) {
  let j = i % 10, k = i % 100;
  if (j === 1 && k !== 11) {
    return i + "st";
  }
  if (j === 2 && k !== 12) {
    return i + "nd";
  }
  if (j === 3 && k !== 13) {
    return i + "rd";
  }
  return i + "th";
}
