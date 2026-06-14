import { ID, idPrefix } from '@monetr/interface/models/ID';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export default class Login {
  readonly [idPrefix] = 'lgn';

  loginId: ID<Login>;
  email: string;
  firstName: string;
  lastName: string;
  totpEnabledAt: Date | null;

  constructor(data: WithJsonValues<Login>) {
    this.loginId = ID.from(data.loginId);
    this.email = data.email;
    this.firstName = data.firstName;
    this.lastName = data.lastName;
    this.totpEnabledAt = parseDate(data.totpEnabledAt);
  }
}
