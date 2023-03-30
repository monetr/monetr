import { RRule } from 'rrule';
import capitalize from 'util/capitalize';

export default class Recurrence {
  name: string;
  rule: RRule;
  ruleText: string;

  constructor(recurrence?: Partial<Recurrence>) {
    if (recurrence) {
      Object.assign(this, recurrence);
    }
    if (this.rule) {
      this.ruleText = capitalize(this.rule.toText());
    }
  }

  ruleString(): string {
    return this.rule.toString().replace('RRULE:', '');
  }
}
