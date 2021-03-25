import moment from "moment";
import { parseToMoment, parseToMomentMaybe } from "util/parseToMoment";

export interface TransactionFields {
  transactionId?: number;
  bankAccountId?: number;
  amount?: number;
  expenseId?: number;
  categories?: string[];
  originalCategories?: string[];
  date?: moment.Moment | string;
  authorizedDate?: moment.Moment | string;
  name?: string;
  originalName?: string;
  merchantName?: string;
  originalMerchantName?: string;
  isPending?: boolean;
  createdAt?: moment.Moment | string;
}


export default class Transaction implements TransactionFields {
  transactionId: number;
  bankAccountId: number;
  amount: number;
  expenseId?: number;
  categories: string[];
  originalCategories: string[];
  date: moment.Moment;
  authorizedDate?: moment.Moment;
  name?: string;
  originalName: string;
  merchantName?: string;
  originalMerchantName?: string;
  isPending: boolean;
  createdAt: moment.Moment;

  constructor(data?: TransactionFields) {
    if (data) {
      Object.assign(this, {
        ...data,
        date: parseToMoment(data.date),
        authorizedDate: parseToMomentMaybe(data.authorizedDate),
        createdAt: parseToMoment(data.createdAt),
      });
    }
  }

  getAmountString(): string {
    if (this.amount < 0) {
      return `+ $${ (-this.amount / 100).toFixed(2) }`
    }

    return `$${ (this.amount / 100).toFixed(2) }`
  }

  getIsAddition(): boolean {
    return this.amount < 0;
  }
}
