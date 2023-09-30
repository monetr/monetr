import { RRule } from 'rrule';

export default class Recurrence {
  name: string;
  dtstart: Date;
  rule: RRule;

  constructor(recurrence?: Partial<Recurrence>) {
    if (recurrence) Object.assign(this, recurrence);
  }

  ruleString(): string {
    return this.rule.toString().replace('RRULE:', '');
  }

  correctRuleString(): string {
    return this.rule.toString();
  }
}
