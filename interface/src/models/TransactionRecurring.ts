import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID, idPrefix } from '@monetr/interface/models/ID';
import Spending from '@monetr/interface/models/Spending';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export enum TransactionRecurringWindow {
  FirstAndFifteenth = 'firstAndFifteenth',
  FifteenthAndLast = 'fifteenthAndLast',
  Weekly = 'weekly',
  BiWeekly = 'biweekly',
  Monthly = 'monthly',
  BiMonthly = 'bimonthly',
  Quarterly = 'quarterly',
  SemiYearly = 'semiyearly',
  Yearly = 'yearly',
}

export default class TransactionRecurring {
  readonly [idPrefix] = 'txrc';

  readonly transactionRecurringId: ID<TransactionRecurring>;
  readonly bankAccountId: ID<BankAccount>;
  readonly transactionClusterId: ID<TransactionCluster>;
  readonly window: TransactionRecurringWindow;
  readonly ruleset: string;
  readonly first: Date;
  readonly last: Date;
  readonly next: Date;
  readonly ended: boolean;
  readonly confidence: number;
  readonly direction: 'debit' | 'credit';
  readonly amounts: { [key: number]: number };
  readonly lastAmount: number;
  // spending is the expense created from this recurring transaction, if there is one
  readonly spending: Spending | null;
  readonly createdAt: Date;
  readonly updatedAt: Date;

  constructor(data: WithJsonValues<TransactionRecurring>) {
    this.transactionRecurringId = ID.from(data.transactionRecurringId);
    this.bankAccountId = ID.from(data.bankAccountId);
    this.transactionClusterId = ID.from(data.transactionClusterId);
    this.window = data.window;
    this.ruleset = data.ruleset;
    this.first = parseDate(data.first);
    this.last = parseDate(data.last);
    this.next = parseDate(data.next);
    this.ended = data.ended;
    this.confidence = data.confidence;
    this.direction = data.direction;
    this.amounts = data.amounts;
    this.lastAmount = data.lastAmount;
    this.spending = data.spending ? new Spending(data.spending) : null;
    this.createdAt = parseDate(data.createdAt);
    this.updatedAt = parseDate(data.updatedAt);
  }
}
