import BankAccount, { type BankAccountSubType, type BankAccountType } from '@monetr/interface/models/BankAccount';

import sortAccounts from './sortAccounts';

describe('sort accounts', () => {
  it('will handle a null or undefined input', () => {
    const foo = sortAccounts(null);
    expect(foo).toEqual([]);

    const bar = sortAccounts(undefined);
    expect(bar).toEqual([]);
  });

  it('will make sure that checking is the highest priority', () => {
    const checkingAccount = new BankAccount({
      bankAccountId: 'abc',
      mask: '1234',
      name: 'Generic Checking',
      accountType: 'depository' as BankAccountType,
      accountSubType: 'checking' as BankAccountSubType,
    });
    const savingsAccount = new BankAccount({
      bankAccountId: 'abd',
      mask: '1234',
      name: 'Generic Savings',
      accountType: 'depository' as BankAccountType,
      accountSubType: 'savings' as BankAccountSubType,
    });
    const accounts = [savingsAccount, checkingAccount];

    const result = sortAccounts(accounts);
    expect(result[0]).toBe(checkingAccount);
    expect(result[1]).toBe(savingsAccount);
  });

  it('will also handle credit card', () => {
    const checkingAccount = new BankAccount({
      bankAccountId: 'abc',
      mask: '1234',
      name: 'Generic Checking',
      accountType: 'depository' as BankAccountType,
      accountSubType: 'checking' as BankAccountSubType,
    });
    const savingsAccount = new BankAccount({
      bankAccountId: 'abd',
      mask: '1234',
      name: 'Generic Savings',
      accountType: 'depository' as BankAccountType,
      accountSubType: 'savings' as BankAccountSubType,
    });
    const creditCard = new BankAccount({
      bankAccountId: 'abe',
      mask: '1234',
      name: 'Generic Credit Card',
      status: 'active',
      accountType: 'credit' as BankAccountType,
      accountSubType: 'credit card' as BankAccountSubType,
    });
    const accounts = [savingsAccount, creditCard, checkingAccount];

    const result = sortAccounts(accounts);
    expect(result[0]).toBe(checkingAccount);
    expect(result[1]).toBe(savingsAccount);
    expect(result[2]).toBe(creditCard);
  });

  it('will put inactive last', () => {
    const checkingAccount = new BankAccount({
      bankAccountId: 'abc',
      mask: '1234',
      name: 'Generic Checking',
      accountType: 'depository' as BankAccountType,
      accountSubType: 'checking' as BankAccountSubType,
    });
    const checkingAccountInactive = new BankAccount({
      bankAccountId: 'abcinactive',
      mask: '1234',
      name: 'Generic Checking',
      accountType: 'depository' as BankAccountType,
      accountSubType: 'checking' as BankAccountSubType,
      status: 'inactive',
    });
    const savingsAccount = new BankAccount({
      bankAccountId: 'abd',
      mask: '1234',
      name: 'Generic Savings',
      accountType: 'depository' as BankAccountType,
      accountSubType: 'savings' as BankAccountSubType,
    });
    const savingsAccountInactive = new BankAccount({
      bankAccountId: 'abdinactive',
      mask: '1234',
      name: 'Generic Savings',
      accountType: 'depository' as BankAccountType,
      accountSubType: 'savings' as BankAccountSubType,
      status: 'inactive',
    });
    const creditCard = new BankAccount({
      bankAccountId: 'abe',
      mask: '1234',
      name: 'Generic Credit Card',
      accountType: 'credit' as BankAccountType,
      accountSubType: 'credit card' as BankAccountSubType,
    });
    const autoInactive = new BankAccount({
      bankAccountId: 'autoinactive',
      mask: '1234',
      name: 'Generic Auto Loan',
      accountType: 'loan' as BankAccountType,
      accountSubType: 'auto' as BankAccountSubType,
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
