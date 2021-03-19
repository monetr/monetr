import { BankAccountFields } from "data/BankAccount";

export interface BalanceFields {
  bankAccountId: number;
  availableBalance: number;
  currentBalance: number;
  safeToSpendBalance: number;
  expensesBalance: number;
  goalsBalance: number;
}

export default class Balance implements BalanceFields {
  bankAccountId: number;
  availableBalance: number;
  currentBalance: number;
  safeToSpendBalance: number;
  expensesBalance: number;
  goalsBalance: number;

  constructor(data?: BankAccountFields) {
    if (data) {
      Object.assign(this, data)
    }
  }
}
