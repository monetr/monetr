import { ID, idPrefix } from '@monetr/interface/models/ID';
import type Link from '@monetr/interface/models/Link';
import Login from '@monetr/interface/models/Login';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export default class User {
  readonly [idPrefix] = 'user';

  readonly userId: ID<User>;
  readonly loginId: string;
  readonly accountId: string;
  readonly account: {
    accountId: string;
    subscriptionActiveUntil: string;
    subscriptionStatus: string;
    timezone: string;
    locale: string;
    createdAt: Date;
  };
  readonly login: Login;
  readonly role: 'member' | 'owner';
  linkOrder: Array<ID<Link>>;

  constructor(data: WithJsonValues<User>) {
    this.userId = ID.from(data.userId);
    this.loginId = data.loginId;
    this.accountId = data.accountId;
    this.account = {
      ...data.account,
      createdAt: parseDate(data.account.createdAt),
    };
    this.login = new Login(data.login);
    this.role = data.role;
    this.linkOrder = data.linkOrder ?? [];
  }

  name(): string {
    return `${this.login.firstName} ${this.login.lastName}`.trim();
  }
}
