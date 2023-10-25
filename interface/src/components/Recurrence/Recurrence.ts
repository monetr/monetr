import { RRule, rrulestr } from 'rrule';

export default class Recurrence {
  name: string;
  rule: RRule;

  constructor(recurrence?: Partial<Recurrence>) {
    if (recurrence) Object.assign(this, recurrence);
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
}
