import type BankAccount from '@monetr/interface/models/BankAccount';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID, idPrefix } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export default class Transaction {
  readonly [idPrefix] = 'txn';

  readonly transactionId: ID<Transaction>;
  readonly bankAccountId: ID<BankAccount>;
  amount: number;
  spendingId: ID<Spending> | null;
  readonly spendingAmount: number | null;
  readonly createdBySpendingId: ID<Spending> | null;
  readonly createdByFundingScheduleId: ID<FundingSchedule> | null;
  readonly categories: string[];
  date: Date;
  authorizedDate: Date | null;
  name: string | null;
  readonly originalName: string;
  merchantName: string | null;
  readonly originalMerchantName: string | null;
  isPending: boolean;
  readonly createdAt: Date;

  constructor(data: WithJsonValues<Transaction>) {
    this.transactionId = ID.from(data.transactionId);
    this.bankAccountId = ID.from(data.bankAccountId);
    this.amount = data.amount;
    this.spendingId = data.spendingId ? ID.from(data.spendingId) : null;
    this.spendingAmount = data.spendingAmount ?? null;
    this.createdBySpendingId = data.createdBySpendingId ? ID.from(data.createdBySpendingId) : null;
    this.createdByFundingScheduleId = data.createdByFundingScheduleId ? ID.from(data.createdByFundingScheduleId) : null;
    this.categories = data.categories ?? [];
    this.date = parseDate(data.date);
    this.authorizedDate = parseDate(data.authorizedDate);
    this.name = data.name ?? null;
    this.originalName = data.originalName;
    this.merchantName = data.merchantName ?? null;
    this.originalMerchantName = data.originalMerchantName ?? null;
    this.isPending = data.isPending;
    this.createdAt = parseDate(data.createdAt);
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
  getMerchantName(): string | null {
    if (this.merchantName) {
      return this.merchantName;
    }

    return this.originalMerchantName ?? null;
  }

  // getMainCategory will return the first category in the categories array. It will first check if a custom category
  // has been specified for the transaction. If there is not one then it will try to use the original categories from
  // the transaction. If those are still not present then it will return "Other" as it cannot infer the transaction's
  // category.
  getMainCategory(): string {
    if (this.categories && this.categories.length > 0) {
      return this.categories[0];
    }

    return 'Other';
  }
}
