import { getRecurringTransactionHistory } from '@monetr/interface/hooks/useRecurringTransactionHistory';
import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import Transaction from '@monetr/interface/models/Transaction';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import TransactionRecurring, { TransactionRecurringWindow } from '@monetr/interface/models/TransactionRecurring';
import type { WithJsonValues } from '@monetr/interface/util/json';

const recurringId = ID.from<TransactionRecurring>('txrc_test');

function recurring(amounts: { [key: number]: number }): TransactionRecurring {
  return new TransactionRecurring({
    transactionRecurringId: recurringId,
    bankAccountId: ID.from<BankAccount>('bac_test'),
    transactionClusterId: ID.from<TransactionCluster>('tcl_test'),
    window: TransactionRecurringWindow.Monthly,
    ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
    first: '2026-01-15T06:00:00Z',
    last: '2026-06-15T05:00:00Z',
    next: '2026-07-15T05:00:00Z',
    ended: false,
    confidence: 0.9,
    direction: 'debit',
    amounts,
    lastAmount: 0,
    spending: null,
    fundingSchedule: null,
    createdAt: '2026-01-15T06:00:00Z',
    updatedAt: '2026-06-16T06:00:00Z',
  });
}

function transaction(amount: number, date: string, overrides: Partial<WithJsonValues<Transaction>> = {}): Transaction {
  return new Transaction({
    transactionId: ID.from<Transaction>(`txn_${date}`),
    bankAccountId: ID.from<BankAccount>('bac_test'),
    amount,
    spendingId: null,
    spendingAmount: null,
    createdBySpendingId: null,
    createdByFundingScheduleId: null,
    categories: [],
    date,
    authorizedDate: null,
    name: 'Netflix',
    originalName: 'NETFLIX.COM',
    merchantName: null,
    originalMerchantName: null,
    isPending: false,
    transactionClusterId: ID.from<TransactionCluster>('tcl_test'),
    transactionRecurringId: recurringId,
    createdAt: date,
    ...overrides,
  });
}

describe('recurring transaction history', () => {
  it('will handle no transactions', () => {
    const result = getRecurringTransactionHistory(recurring({ 1549: 3 }), []);
    expect(result.transactions).toEqual([]);
    expect(result.seen).toBe(3);
    expect(result.priceChange).toBeNull();
  });

  it('will add up seen from all of the amounts', () => {
    const result = getRecurringTransactionHistory(recurring({ 1399: 3, 1549: 5, 1699: 1 }), []);
    expect(result.seen).toBe(9);
  });

  it('will not have a price change when the amount never changed', () => {
    const result = getRecurringTransactionHistory(recurring({ 1549: 3 }), [
      transaction(1549, '2026-03-15T05:00:00Z'),
      transaction(1549, '2026-02-15T06:00:00Z'),
      transaction(1549, '2026-01-15T06:00:00Z'),
    ]);
    expect(result.transactions).toHaveLength(3);
    expect(result.priceChange).toBeNull();
  });

  it('will find a price increase', () => {
    const result = getRecurringTransactionHistory(recurring({ 1399: 2, 1549: 2 }), [
      transaction(1549, '2026-04-15T05:00:00Z'),
      transaction(1549, '2026-03-15T05:00:00Z'),
      transaction(1399, '2026-02-15T06:00:00Z'),
      transaction(1399, '2026-01-15T06:00:00Z'),
    ]);
    expect(result.priceChange).toEqual({
      previousAmount: 1399,
      previousCount: 2,
      currentAmount: 1549,
      currentCount: 2,
      changedAt: new Date('2026-03-15T05:00:00Z'),
    });
  });

  it('will find a price decrease', () => {
    const result = getRecurringTransactionHistory(recurring({ 1549: 1, 999: 1 }), [
      transaction(999, '2026-02-15T06:00:00Z'),
      transaction(1549, '2026-01-15T06:00:00Z'),
    ]);
    expect(result.priceChange?.previousAmount).toBe(1549);
    expect(result.priceChange?.currentAmount).toBe(999);
    expect(result.priceChange?.changedAt).toEqual(new Date('2026-02-15T06:00:00Z'));
  });

  it('will only report the most recent price change', () => {
    const result = getRecurringTransactionHistory(recurring({ 1299: 1, 1399: 2, 1549: 1 }), [
      transaction(1549, '2026-04-15T05:00:00Z'),
      transaction(1399, '2026-03-15T05:00:00Z'),
      transaction(1399, '2026-02-15T06:00:00Z'),
      transaction(1299, '2026-01-15T06:00:00Z'),
    ]);
    expect(result.priceChange?.previousAmount).toBe(1399);
    expect(result.priceChange?.previousCount).toBe(2);
    expect(result.priceChange?.currentAmount).toBe(1549);
    expect(result.priceChange?.changedAt).toEqual(new Date('2026-04-15T05:00:00Z'));
  });

  it('will fall back to a count of zero for amounts that are not in the amounts', () => {
    const result = getRecurringTransactionHistory(recurring({ 1549: 1 }), [
      transaction(1549, '2026-02-15T06:00:00Z'),
      transaction(1399, '2026-01-15T06:00:00Z'),
    ]);
    expect(result.priceChange?.previousAmount).toBe(1399);
    expect(result.priceChange?.previousCount).toBe(0);
    expect(result.priceChange?.currentCount).toBe(1);
  });
});
