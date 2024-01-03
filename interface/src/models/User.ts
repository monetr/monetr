import Login from '@monetr/interface/models/Login';

export default class User {
  userId: number;
  loginId: number;
  accountId: number;
  account: {
    accountId: number;
    subscriptionActiveUntil: string;
    subscriptionStatus: string;
    timezone: string;
  };
  login: Login;

  constructor(data?: Partial<User>) {
    if (data) Object.assign(this, {
      ...data,
      login: new Login(data?.login),
    });
  }
}
