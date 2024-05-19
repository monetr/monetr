export default class Login {
  loginId: string;
  email: string;
  firstName: string;
  lastName: string;

  constructor(data?: Partial<Login>) {
    if (data) Object.assign(this, data);
  }
};
