import moment from "moment";
import { parseToMoment, parseToMomentMaybe } from "util/parseToMoment";

export default class Transaction {
  transactionId: number;
  bankAccountId: number;
  amount: number;
  spendingId?: number;
  spendingAmount?: number;
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

  constructor(data?: Partial<Transaction>) {
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

  getName(): string {
    if (this.name) {
      return this.name;
    }

    return this.originalName;
  }

  getOriginalName(): string {
    return this.originalName;
  }

  // getMerchantName will return the custom merchant name specified by the user (if there is one) or it will return the
  // original merchant name from when the transaction was initially created. Transaction's do not require a merchant
  // name at all though so this may still return null.
  getMerchantName(): string|null {
    if (this.merchantName) {
      return this.merchantName;
    }

    return this.originalMerchantName;
  }

  // getMainCategory will return the first category in the categories array. It will first check if a custom category
  // has been specified for the transaction. If there is not one then it will try to use the original categories from
  // the transaction. If those are still not present then it will return "Other" as it cannot infer the transaction's
  // category.
  getMainCategory(): string {
    if (this.categories && this.categories.length > 0) {
      return this.categories[0];
    }

    if (this.originalCategories && this.originalCategories.length > 0) {
      return this.originalCategories[0];
    }

    return "Other";
  }
}
