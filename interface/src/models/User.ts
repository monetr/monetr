import Login from '@monetr/interface/models/Login';

export default class User {
  userId: string;
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
    if (data)
      Object.assign(this, {
        ...data,
        login: new Login(data?.login),
      });
  }
}
