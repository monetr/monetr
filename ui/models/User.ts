import Login from 'models/Login';

export default class User {
  userId: number;
  loginId: number;
  accountId: number;
  login: Login;
  firstName: string;
  lastName: string;

  constructor(data?: Partial<User>) {
    if (data) {
      Object.assign(this, {
        ...data,
        login: new Login(data?.login),
      });
    }
  }
}
