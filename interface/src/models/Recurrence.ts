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
      // rrule types these option fields as non-nullable, but we deliberately null them out so they're ignored when we
      // compare the two rules. Alias the options through a loose view so the assignments are allowed.
      const inputOptions = inputRule.options as unknown as Record<string, unknown>;
      inputOptions.dtstart = null;
      inputOptions.byhour = null;
      inputOptions.byminute = null;
      inputOptions.bysecond = null;
      inputOptions.tzid = null;

      const thisRule = this.rule.clone();
      const thisOptions = thisRule.options as unknown as Record<string, unknown>;
      thisOptions.dtstart = null;
      thisOptions.byhour = null;
      thisOptions.byminute = null;
      thisOptions.bysecond = null;
      inputOptions.tzid = null;

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
      // rrule types these option fields as non-nullable, but we deliberately null them out so they're ignored when we
      // compare the two rules. Alias the options through a loose view so the assignments are allowed.
      const inputOptions = inputRule.options as unknown as Record<string, unknown>;
      inputOptions.dtstart = null;
      inputOptions.byhour = null;
      inputOptions.byminute = null;
      inputOptions.bysecond = null;
      inputOptions.tzid = null;

      const thisRule = this.rule.clone();
      const thisOptions = thisRule.options as unknown as Record<string, unknown>;
      thisOptions.dtstart = null;
      thisOptions.byhour = null;
      thisOptions.byminute = null;
      thisOptions.bysecond = null;
      inputOptions.tzid = null;

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
