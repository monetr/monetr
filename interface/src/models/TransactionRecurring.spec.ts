import type BankAccount from '@monetr/interface/models/BankAccount';
import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID } from '@monetr/interface/models/ID';
import Spending, { SpendingType } from '@monetr/interface/models/Spending';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import TransactionRecurring, { TransactionRecurringWindow } from '@monetr/interface/models/TransactionRecurring';
import type { WithJsonValues } from '@monetr/interface/util/json';

function fixture(overrides: Partial<WithJsonValues<TransactionRecurring>> = {}): TransactionRecurring {
  return new TransactionRecurring({
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
    spending: null,
    fundingSchedule: null,
    transactionCluster: null,
    createdAt: '2026-03-15T06:00:00Z',
    updatedAt: '2026-03-16T06:00:00Z',
    ...overrides,
  });
}

describe('transaction recurring', () => {
  it('will parse the ids', () => {
    const recurring = fixture();
    expect(recurring.transactionRecurringId.toString()).toBe('txrc_test');
    expect(recurring.bankAccountId.toString()).toBe('bac_test');
    expect(recurring.transactionClusterId.toString()).toBe('tcl_test');
  });

  it('will parse the dates', () => {
    const recurring = fixture();
    expect(recurring.first).toEqual(new Date('2026-01-15T06:00:00Z'));
    expect(recurring.last).toEqual(new Date('2026-03-15T05:00:00Z'));
    expect(recurring.next).toEqual(new Date('2026-04-15T05:00:00Z'));
    expect(recurring.createdAt).toEqual(new Date('2026-03-15T06:00:00Z'));
    expect(recurring.updatedAt).toEqual(new Date('2026-03-16T06:00:00Z'));
  });

  it('will keep the window, direction and amounts', () => {
    const recurring = fixture({
      window: TransactionRecurringWindow.FifteenthAndLast,
      direction: 'credit',
      amounts: {
        '-500000': 4,
      },
      lastAmount: -500000,
    });
    expect(recurring.window).toBe(TransactionRecurringWindow.FifteenthAndLast);
    expect(recurring.direction).toBe('credit');
    expect(recurring.amounts).toEqual({
      '-500000': 4,
    });
    expect(recurring.lastAmount).toBe(-500000);
  });

  it('will not have spending when nothing was created from it', () => {
    const recurring = fixture();
    expect(recurring.spending).toBeNull();
  });

  it('will parse the spending created from it', () => {
    const recurring = fixture({
      spendingId: ID.from<Spending>('spnd_test'),
      spending: {
        spendingId: ID.from<Spending>('spnd_test'),
        bankAccountId: ID.from<BankAccount>('bac_test'),
        fundingScheduleId: ID.from('fund_test'),
        name: 'Github',
        spendingType: SpendingType.Expense,
        targetAmount: 800,
        currentAmount: 0,
        usedAmount: 0,
        ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
        lastRecurrence: null,
        nextRecurrence: '2026-04-15T05:00:00Z',
        nextContributionAmount: 400,
        isBehind: false,
        isPaused: false,
        autoCreateTransaction: false,
        createdAt: '2026-03-15T06:00:00Z',
      },
    });
    expect(recurring.spendingId?.toString()).toBe('spnd_test');
    expect(recurring.spending).toBeInstanceOf(Spending);
    expect(recurring.spending?.spendingId.toString()).toBe('spnd_test');
    expect(recurring.spending?.nextRecurrence).toEqual(new Date('2026-04-15T05:00:00Z'));
  });

  it('will not have a funding schedule when nothing was created from it', () => {
    const recurring = fixture();
    expect(recurring.fundingSchedule).toBeNull();
  });

  it('will parse the funding schedule created from it', () => {
    const recurring = fixture({
      direction: 'credit',
      fundingScheduleId: ID.from<FundingSchedule>('fund_test'),
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
    });
    expect(recurring.fundingScheduleId?.toString()).toBe('fund_test');
    expect(recurring.fundingSchedule).toBeInstanceOf(FundingSchedule);
    expect(recurring.fundingSchedule?.fundingScheduleId.toString()).toBe('fund_test');
    expect(recurring.fundingSchedule?.nextRecurrence).toEqual(new Date('2026-04-15T05:00:00Z'));
  });

  it('will throw on an invalid date', () => {
    expect(() =>
      fixture({
        next: 'not a date',
      }),
    ).toThrow();
  });
});
