import moment from "moment";

export interface TransactionFields {
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

  constructor(data: TransactionFields) {
    Object.assign(this, data);
  }
}
