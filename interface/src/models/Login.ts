import { ID, idPrefix } from '@monetr/interface/models/ID';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export default class Login {
  readonly [idPrefix] = 'lgn';

  loginId: ID<Login>;
  email: string;
  firstName: string;
  lastName: string;
  isEmailVerified: boolean;
  emailVerifiedAt: Date | null;
  totpEnabledAt: Date | null;

  constructor(data: WithJsonValues<Login>) {
    this.loginId = ID.from(data.loginId);
    this.email = data.email;
    this.firstName = data.firstName;
    this.lastName = data.lastName;
    this.isEmailVerified = data.isEmailVerified;
    this.emailVerifiedAt = parseDate(data.emailVerifiedAt);
    this.totpEnabledAt = parseDate(data.totpEnabledAt);
  }

  // Same as GetEmailIsVerified on the server
  getIsEmailVerified(): boolean {
    return this.isEmailVerified && this.emailVerifiedAt !== null;
  }
}
