import { ID, idPrefix } from '@monetr/interface/models/ID';
import Login from '@monetr/interface/models/Login';
import type { WithJsonValues } from '@monetr/interface/util/json';

export default class User {
  readonly [idPrefix] = 'user';

  userId: ID<User>;
  loginId: string;
  accountId: string;
  account: {
    accountId: string;
    subscriptionActiveUntil: string;
    subscriptionStatus: string;
    timezone: string;
    locale: string;
  };
  login: Login;

  constructor(data: WithJsonValues<User>) {
    this.userId = ID.from(data.userId);
    this.loginId = data.loginId;
    this.accountId = data.accountId;
    this.account = data.account;
    this.login = new Login(data.login);
  }
}
