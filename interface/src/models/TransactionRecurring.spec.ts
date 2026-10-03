import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import TransactionRecurring, { TransactionRecurringWindow } from '@monetr/interface/models/TransactionRecurring';
import type { WithJsonValues } from '@monetr/interface/util/json';

function fixture(overrides: Partial<WithJsonValues<TransactionRecurring>> = {}): TransactionRecurring {
  return new TransactionRecurring({
    transactionRecurringId: ID.from<TransactionRecurring>('txrc_test'),
    bankAccountId: ID.from<BankAccount>('bac_test'),
    transactionClusterId: ID.from<TransactionCluster>('tcl_test'),
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

  it('will throw on an invalid date', () => {
    expect(() =>
      fixture({
        next: 'not a date',
      }),
    ).toThrow();
  });
});
