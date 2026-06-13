import BankAccount, { BankAccountSubType, BankAccountType } from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type Link from '@monetr/interface/models/Link';
import type { WithJsonValues } from '@monetr/interface/util/json';

import sortAccounts from './sortAccounts';

// The strict BankAccount constructor wants the entire JSON shape now, but these sorting tests only care about a handful
// of fields. This helper fills in sensible defaults so each case can just override the bits that matter to it. The ids
// here are made up so we use the two type arg form of ID.from to skip the prefix template check.
function fixture(overrides: Partial<WithJsonValues<BankAccount>>): BankAccount {
  return new BankAccount({
    bankAccountId: ID.from<BankAccount, string>('abc'),
    linkId: ID.from<Link, string>('link_test'),
    lunchFlowBankAccountId: null,
    mask: '1234',
    name: 'Generic Account',
    originalName: 'Generic Account',
    status: 'active',
    accountType: BankAccountType.Depository,
    accountSubType: BankAccountSubType.Checking,
    currency: 'USD',
    currentBalance: 0,
    availableBalance: 0,
    limitBalance: null,
    lastUpdated: new Date(),
    createdAt: new Date(),
    createdBy: 'user_test',
    deletedAt: null,
    plaidBankAccount: null,
    lunchFlowBankAccount: null,
    ...overrides,
  });
}

describe('sort accounts', () => {
  it('will handle a null or undefined input', () => {
    const foo = sortAccounts(null);
    expect(foo).toEqual([]);

    const bar = sortAccounts(undefined);
    expect(bar).toEqual([]);
  });

  it('will make sure that checking is the highest priority', () => {
    const checkingAccount = fixture({
      bankAccountId: ID.from<BankAccount, string>('abc'),
      mask: '1234',
      name: 'Generic Checking',
      accountType: BankAccountType.Depository,
      accountSubType: BankAccountSubType.Checking,
    });
    const savingsAccount = fixture({
      bankAccountId: ID.from<BankAccount, string>('abd'),
      mask: '1234',
      name: 'Generic Savings',
      accountType: BankAccountType.Depository,
      accountSubType: BankAccountSubType.Savings,
    });
    const accounts = [savingsAccount, checkingAccount];

    const result = sortAccounts(accounts);
    expect(result[0]).toBe(checkingAccount);
    expect(result[1]).toBe(savingsAccount);
  });

  it('will also handle credit card', () => {
    const checkingAccount = fixture({
      bankAccountId: ID.from<BankAccount, string>('abc'),
      mask: '1234',
      name: 'Generic Checking',
      accountType: BankAccountType.Depository,
      accountSubType: BankAccountSubType.Checking,
    });
    const savingsAccount = fixture({
      bankAccountId: ID.from<BankAccount, string>('abd'),
      mask: '1234',
      name: 'Generic Savings',
      accountType: BankAccountType.Depository,
      accountSubType: BankAccountSubType.Savings,
    });
    const creditCard = fixture({
      bankAccountId: ID.from<BankAccount, string>('abe'),
      mask: '1234',
      name: 'Generic Credit Card',
      status: 'active',
      accountType: BankAccountType.Credit,
      accountSubType: BankAccountSubType.CreditCard,
    });
    const accounts = [savingsAccount, creditCard, checkingAccount];

    const result = sortAccounts(accounts);
    expect(result[0]).toBe(checkingAccount);
    expect(result[1]).toBe(savingsAccount);
    expect(result[2]).toBe(creditCard);
  });

  it('will put inactive last', () => {
    const checkingAccount = fixture({
      bankAccountId: ID.from<BankAccount, string>('abc'),
      mask: '1234',
      name: 'Generic Checking',
      accountType: BankAccountType.Depository,
      accountSubType: BankAccountSubType.Checking,
    });
    const checkingAccountInactive = fixture({
      bankAccountId: ID.from<BankAccount, string>('abcinactive'),
      mask: '1234',
      name: 'Generic Checking',
      accountType: BankAccountType.Depository,
      accountSubType: BankAccountSubType.Checking,
      status: 'inactive',
    });
    const savingsAccount = fixture({
      bankAccountId: ID.from<BankAccount, string>('abd'),
      mask: '1234',
      name: 'Generic Savings',
      accountType: BankAccountType.Depository,
      accountSubType: BankAccountSubType.Savings,
    });
    const savingsAccountInactive = fixture({
      bankAccountId: ID.from<BankAccount, string>('abdinactive'),
      mask: '1234',
      name: 'Generic Savings',
      accountType: BankAccountType.Depository,
      accountSubType: BankAccountSubType.Savings,
      status: 'inactive',
    });
    const creditCard = fixture({
      bankAccountId: ID.from<BankAccount, string>('abe'),
      mask: '1234',
      name: 'Generic Credit Card',
      accountType: BankAccountType.Credit,
      accountSubType: BankAccountSubType.CreditCard,
    });
    const autoInactive = fixture({
      bankAccountId: ID.from<BankAccount, string>('autoinactive'),
      mask: '1234',
      name: 'Generic Auto Loan',
      accountType: BankAccountType.Loan,
      accountSubType: BankAccountSubType.Auto,
      status: 'inactive',
    });
    const accounts = [
      autoInactive,
      checkingAccount,
      checkingAccountInactive,
      creditCard,
      savingsAccount,
      savingsAccountInactive,
    ];

    const result = sortAccounts(accounts);
    expect(result[0]).toBe(checkingAccount);
    expect(result[1]).toBe(savingsAccount);
    expect(result[2]).toBe(creditCard);
    expect(result[3]).toBe(autoInactive);
    expect(result[4]).toBe(savingsAccountInactive);
    expect(result[5]).toBe(checkingAccountInactive);
  });
});
