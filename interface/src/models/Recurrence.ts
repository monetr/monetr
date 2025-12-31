import { type RRule, rrulestr } from 'rrule';

export default class Recurrence {
  name: string;
  rule: RRule;

  constructor(recurrence?: Partial<Recurrence>) {
    if (recurrence) {
      Object.assign(this, recurrence);
    }
  }

  ruleString(): string {
    return this.rule.toString();
  }

  equalRule(input: string): boolean {
    try {
      const inputRule = rrulestr(input);
      inputRule.options.dtstart = null;
      inputRule.options.byhour = null;
      inputRule.options.byminute = null;
      inputRule.options.bysecond = null;
      inputRule.options.tzid = null;

      const thisRule = this.rule.clone();
      thisRule.options.dtstart = null;
      thisRule.options.byhour = null;
      thisRule.options.byminute = null;
      thisRule.options.bysecond = null;
      inputRule.options.tzid = null;

      const a = JSON.stringify(inputRule.options);
      const b = JSON.stringify(thisRule.options);

      return a === b;
    } catch {
      console.warn('cannot compare invalid rrules', input, this);
      return false;
    }
  }

  /*
   * signature returns a string representing a soft way of identifying a recurrence rule in a dropdown list.
   */
  signature(): string {
    return ruleSignature(this.rule);
  }

  equalSignature(input: string): boolean {
    try {
      const inputRule = rrulestr(input);
      inputRule.options.dtstart = null;
      inputRule.options.byhour = null;
      inputRule.options.byminute = null;
      inputRule.options.bysecond = null;
      inputRule.options.tzid = null;

      const thisRule = this.rule.clone();
      thisRule.options.dtstart = null;
      thisRule.options.byhour = null;
      thisRule.options.byminute = null;
      thisRule.options.bysecond = null;
      inputRule.options.tzid = null;

      const a = ruleSignature(inputRule);
      const b = ruleSignature(thisRule);

      return a === b;
    } catch {
      console.warn('cannot compare invalid rrules', input, this);
      return false;
    }
  }
}

export function ruleSignature(rule: RRule): string {
  if (Array.isArray(rule.origOptions.bymonthday) && rule.origOptions.bymonthday.length > 1) {
    return `${rule.options.freq}::${rule.options.interval}::${rule.origOptions.bymonthday}`;
  }

  return `${rule.options.freq}::${rule.options.interval}`;
}
