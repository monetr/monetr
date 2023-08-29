import { parseJSON } from 'date-fns';
import formatAmount from 'util/formatAmount';

export default class Transaction {
  transactionId: number;
  bankAccountId: number;
  amount: number;
  spendingId?: number;
  spendingAmount?: number;
  categories: string[];
  originalCategories: string[];
  date: Date;
  authorizedDate?: Date;
  name?: string;
  originalName: string;
  merchantName?: string;
  originalMerchantName?: string;
  isPending: boolean;
  createdAt: Date;

  constructor(data?: Partial<Transaction>) {
    if (data) {
      Object.assign(this, {
        ...data,
        date: parseJSON(data.date),
        authorizedDate: data.authorizedDate ?? parseJSON(data.authorizedDate),
        createdAt: parseJSON(data.createdAt),
      });
    }
  }

  getAmountString(): string {
    const amount = Math.abs(this.amount);
    if (this.amount < 0) {
      return `+ ${ formatAmount(amount) }`;
    }

    return formatAmount(amount);
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

  // getTitle will return a transaction title that can be used in listing transactions. This is meant to return the most
  // friendly title first, but fallback on other data if a friendly title is not available.
  getTitle(): string {
    if (this.name) {
      return this.name;
    } else if (this.merchantName) {
      return this.merchantName;
    } else if (this.originalMerchantName) {
      return this.originalMerchantName;
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

    return 'Other';
  }
}
