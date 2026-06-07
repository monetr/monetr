import { ID, idPrefix } from '@monetr/interface/models/ID';
import Login from '@monetr/interface/models/Login';

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

  constructor(data?: Partial<User>) {
    if (data) {
      Object.assign(this, {
        ...data,
        login: new Login(data?.login),
      });
    }
  }
}
