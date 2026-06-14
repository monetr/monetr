import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type { WithJsonValues } from '@monetr/interface/util/json';

export default class Balance {
  bankAccountId: ID<BankAccount>;
  available: number;
  current: number;
  limit: number;
  free: number;
  expenses: number;
  goals: number;

  constructor(data: WithJsonValues<Balance>) {
    this.bankAccountId = ID.from(data.bankAccountId);
    this.available = data.available;
    this.current = data.current;
    this.limit = data.limit;
    this.free = data.free;
    this.expenses = data.expenses;
    this.goals = data.goals;
  }
}
