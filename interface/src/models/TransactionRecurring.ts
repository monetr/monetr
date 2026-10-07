import type BankAccount from '@monetr/interface/models/BankAccount';
import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID, idPrefix } from '@monetr/interface/models/ID';
import Spending from '@monetr/interface/models/Spending';
import TransactionCluster from '@monetr/interface/models/TransactionCluster';
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
  // spendingId is the expense tracking this recurring transaction, only ever set on debits
  spendingId: ID<Spending> | null;
  // fundingScheduleId is the funding schedule tracking this recurring transaction, only ever set on credits
  fundingScheduleId: ID<FundingSchedule> | null;
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
  // autoMatched is true when monetr linked the expense itself instead of the user picking it
  readonly autoMatched: boolean;
  // spending is the expense tracking this recurring transaction, if there is one
  readonly spending: Spending | null;
  // fundingSchedule is the funding schedule tracking this recurring transaction, if there is one
  readonly fundingSchedule: FundingSchedule | null;
  // transactionCluster is the similar transactions group this was detected in, only included when listing them
  readonly transactionCluster: TransactionCluster | null;
  readonly createdAt: Date;
  readonly updatedAt: Date;

  constructor(data: WithJsonValues<TransactionRecurring>) {
    this.transactionRecurringId = ID.from(data.transactionRecurringId);
    this.bankAccountId = ID.from(data.bankAccountId);
    this.transactionClusterId = ID.from(data.transactionClusterId);
    this.spendingId = data.spendingId ? ID.from(data.spendingId) : null;
    this.fundingScheduleId = data.fundingScheduleId ? ID.from(data.fundingScheduleId) : null;
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
    this.autoMatched = data.autoMatched;
    this.spending = data.spending ? new Spending(data.spending) : null;
    this.fundingSchedule = data.fundingSchedule ? new FundingSchedule(data.fundingSchedule) : null;
    this.transactionCluster = data.transactionCluster ? new TransactionCluster(data.transactionCluster) : null;
    this.createdAt = parseDate(data.createdAt);
    this.updatedAt = parseDate(data.updatedAt);
  }
}
