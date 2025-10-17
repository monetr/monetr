export default class Balance {
  bankAccountId: string;
  available: number;
  current: number;
  limit: number;
  free: number;
  expenses: number;
  goals: number;

  constructor(data?: Partial<Balance>) {
    if (data) { Object.assign(this, data); }
  }
}
