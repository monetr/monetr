import { RRule } from 'rrule';

export default class Recurrence {
  name: string;
  rule: RRule;

  constructor(recurrence?: Partial<Recurrence>) {
    if (recurrence) {
      Object.assign(this, recurrence);
    }
  }

  ruleString(): string {
    return this.rule.toString().replace('RRULE:', '');
  }
}
