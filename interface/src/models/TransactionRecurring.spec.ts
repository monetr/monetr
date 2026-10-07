import type BankAccount from '@monetr/interface/models/BankAccount';
import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import TransactionRecurring, { TransactionRecurringWindow } from '@monetr/interface/models/TransactionRecurring';

describe('transaction recurring', () => {
  it('will parse the ids', () => {
    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_test'),
      bankAccountId: ID.from<BankAccount>('bac_test'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_test'),
      spendingId: null,
      fundingScheduleId: null,
      window: TransactionRecurringWindow.Monthly,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-15T05:00:00Z',
      next: '2026-04-15T05:00:00Z',
      ended: false,
      confidence: 0.9,
      direction: 'debit',
      amounts: {
        800: 3,
      },
      lastAmount: 800,
      autoMatched: false,
      fundingSchedule: null,
      transactionCluster: null,
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });
    expect(recurring.transactionRecurringId).toBe('txrc_test');
    expect(recurring.bankAccountId).toBe('bac_test');
    expect(recurring.transactionClusterId).toBe('tcl_test');
  });

  it('will parse the dates', () => {
    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_test'),
      bankAccountId: ID.from<BankAccount>('bac_test'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_test'),
      spendingId: null,
      fundingScheduleId: null,
      window: TransactionRecurringWindow.Monthly,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-15T05:00:00Z',
      next: '2026-04-15T05:00:00Z',
      ended: false,
      confidence: 0.9,
      direction: 'debit',
      amounts: {
        800: 3,
      },
      lastAmount: 800,
      autoMatched: false,
      fundingSchedule: null,
      transactionCluster: null,
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });
    expect(recurring.first).toEqual(new Date('2026-01-15T06:00:00Z'));
    expect(recurring.last).toEqual(new Date('2026-03-15T05:00:00Z'));
    expect(recurring.next).toEqual(new Date('2026-04-15T05:00:00Z'));
    expect(recurring.createdAt).toEqual(new Date('2026-03-15T06:00:00Z'));
    expect(recurring.updatedAt).toEqual(new Date('2026-03-16T06:00:00Z'));
  });

  it('will keep the window, direction and amounts', () => {
    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_test'),
      bankAccountId: ID.from<BankAccount>('bac_test'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_test'),
      spendingId: null,
      fundingScheduleId: null,
      window: TransactionRecurringWindow.FifteenthAndLast,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-15T05:00:00Z',
      next: '2026-04-15T05:00:00Z',
      ended: false,
      confidence: 0.9,
      direction: 'credit',
      amounts: {
        '-500000': 4,
      },
      lastAmount: -500000,
      autoMatched: false,
      fundingSchedule: null,
      transactionCluster: null,
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });
    expect(recurring.window).toBe(TransactionRecurringWindow.FifteenthAndLast);
    expect(recurring.direction).toBe('credit');
    expect(recurring.amounts).toEqual({
      '-500000': 4,
    });
    expect(recurring.lastAmount).toBe(-500000);
  });

  it('will not have a spending id when nothing is linked', () => {
    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_test'),
      bankAccountId: ID.from<BankAccount>('bac_test'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_test'),
      spendingId: null,
      fundingScheduleId: null,
      window: TransactionRecurringWindow.Monthly,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-15T05:00:00Z',
      next: '2026-04-15T05:00:00Z',
      ended: false,
      confidence: 0.9,
      direction: 'debit',
      amounts: {
        800: 3,
      },
      lastAmount: 800,
      autoMatched: false,
      fundingSchedule: null,
      transactionCluster: null,
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });
    expect(recurring.spendingId).toBeNull();
  });

  it('will parse the linked spending id', () => {
    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_test'),
      bankAccountId: ID.from<BankAccount>('bac_test'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_test'),
      spendingId: ID.from<Spending>('spnd_test'),
      fundingScheduleId: null,
      window: TransactionRecurringWindow.Monthly,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-15T05:00:00Z',
      next: '2026-04-15T05:00:00Z',
      ended: false,
      confidence: 0.9,
      direction: 'debit',
      amounts: {
        800: 3,
      },
      lastAmount: 800,
      autoMatched: false,
      fundingSchedule: null,
      transactionCluster: null,
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });
    expect(recurring.spendingId).toBe('spnd_test');
  });

  it('will describe the confidence in words', () => {
    const veryLikely = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_test'),
      bankAccountId: ID.from<BankAccount>('bac_test'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_test'),
      spendingId: null,
      fundingScheduleId: null,
      window: TransactionRecurringWindow.Monthly,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-15T05:00:00Z',
      next: '2026-04-15T05:00:00Z',
      ended: false,
      confidence: 0.95,
      direction: 'debit',
      amounts: {
        800: 3,
      },
      lastAmount: 800,
      autoMatched: false,
      fundingSchedule: null,
      transactionCluster: null,
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });
    expect(veryLikely.getConfidenceLabel()).toBe('Very likely');
    const likely = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_test'),
      bankAccountId: ID.from<BankAccount>('bac_test'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_test'),
      spendingId: null,
      fundingScheduleId: null,
      window: TransactionRecurringWindow.Monthly,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-15T05:00:00Z',
      next: '2026-04-15T05:00:00Z',
      ended: false,
      confidence: 0.8,
      direction: 'debit',
      amounts: {
        800: 3,
      },
      lastAmount: 800,
      autoMatched: false,
      fundingSchedule: null,
      transactionCluster: null,
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });
    expect(likely.getConfidenceLabel()).toBe('Likely');
    const possibly = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_test'),
      bankAccountId: ID.from<BankAccount>('bac_test'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_test'),
      spendingId: null,
      fundingScheduleId: null,
      window: TransactionRecurringWindow.Monthly,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-15T05:00:00Z',
      next: '2026-04-15T05:00:00Z',
      ended: false,
      confidence: 0.5,
      direction: 'debit',
      amounts: {
        800: 3,
      },
      lastAmount: 800,
      autoMatched: false,
      fundingSchedule: null,
      transactionCluster: null,
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });
    expect(possibly.getConfidenceLabel()).toBe('Possibly');
  });

  it('will not have a funding schedule when nothing was created from it', () => {
    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_test'),
      bankAccountId: ID.from<BankAccount>('bac_test'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_test'),
      spendingId: null,
      fundingScheduleId: null,
      window: TransactionRecurringWindow.Monthly,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-15T05:00:00Z',
      next: '2026-04-15T05:00:00Z',
      ended: false,
      confidence: 0.9,
      direction: 'debit',
      amounts: {
        800: 3,
      },
      lastAmount: 800,
      autoMatched: false,
      fundingSchedule: null,
      transactionCluster: null,
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });
    expect(recurring.fundingSchedule).toBeNull();
  });

  it('will parse the funding schedule created from it', () => {
    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_test'),
      bankAccountId: ID.from<BankAccount>('bac_test'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_test'),
      spendingId: null,
      fundingScheduleId: ID.from<FundingSchedule>('fund_test'),
      window: TransactionRecurringWindow.Monthly,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-15T05:00:00Z',
      next: '2026-04-15T05:00:00Z',
      ended: false,
      confidence: 0.9,
      direction: 'credit',
      amounts: {
        800: 3,
      },
      lastAmount: 800,
      autoMatched: false,
      fundingSchedule: {
        fundingScheduleId: ID.from<FundingSchedule>('fund_test'),
        bankAccountId: ID.from<BankAccount>('bac_test'),
        name: 'Payday',
        description: null,
        ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
        lastRecurrence: null,
        nextRecurrence: '2026-04-15T05:00:00Z',
        nextRecurrenceOriginal: '2026-04-15T05:00:00Z',
        excludeWeekends: false,
        autoCreateTransaction: false,
        estimatedDeposit: 250000,
      },
      transactionCluster: null,
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });
    expect(recurring.fundingScheduleId).toBe('fund_test');
    expect(recurring.fundingSchedule).toBeInstanceOf(FundingSchedule);
    expect(recurring.fundingSchedule?.fundingScheduleId).toBe('fund_test');
    expect(recurring.fundingSchedule?.nextRecurrence).toEqual(new Date('2026-04-15T05:00:00Z'));
  });

  it('will throw on an invalid date', () => {
    expect(
      () =>
        new TransactionRecurring({
          transactionRecurringId: ID.from<TransactionRecurring>('txrc_test'),
          bankAccountId: ID.from<BankAccount>('bac_test'),
          transactionClusterId: ID.from<TransactionCluster>('tcl_test'),
          spendingId: null,
          fundingScheduleId: null,
          window: TransactionRecurringWindow.Monthly,
          ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
          first: '2026-01-15T06:00:00Z',
          last: '2026-03-15T05:00:00Z',
          next: 'not a date',
          ended: false,
          confidence: 0.9,
          direction: 'debit',
          amounts: {
            800: 3,
          },
          lastAmount: 800,
          autoMatched: false,
          fundingSchedule: null,
          transactionCluster: null,
          createdAt: '2026-03-15T06:00:00Z',
          updatedAt: '2026-03-16T06:00:00Z',
        }),
    ).toThrow();
  });
});
