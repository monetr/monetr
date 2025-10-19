import BankAccount from '@monetr/interface/models/BankAccount';

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
      availableBalance: 1234,
      currentBalance: 1234,
      mask: '1234',
      name: 'Generic Checking',
      accountType: 'depository',
      accountSubType: 'checking',
    });
    const savingsAccount = new BankAccount({
      bankAccountId: 'abd',
      availableBalance: 1234,
      currentBalance: 1234,
      mask: '1234',
      name: 'Generic Savings',
      accountType: 'depository',
      accountSubType: 'savings',
    });
    const accounts = [savingsAccount, checkingAccount];

    const result = sortAccounts(accounts);
    expect(result[0]).toBe(checkingAccount);
    expect(result[1]).toBe(savingsAccount);
  });

  it('will also handle credit card', () => {
    const checkingAccount = new BankAccount({
      bankAccountId: 'abc',
      availableBalance: 1234,
      currentBalance: 1234,
      mask: '1234',
      name: 'Generic Checking',
      accountType: 'depository',
      accountSubType: 'checking',
    });
    const savingsAccount = new BankAccount({
      bankAccountId: 'abd',
      availableBalance: 1234,
      currentBalance: 1234,
      mask: '1234',
      name: 'Generic Savings',
      accountType: 'depository',
      accountSubType: 'savings',
    });
    const creditCard = new BankAccount({
      bankAccountId: 'abe',
      availableBalance: 1234,
      currentBalance: 1234,
      mask: '1234',
      name: 'Generic Credit Card',
      accountType: 'credit',
      accountSubType: 'credit card',
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
      availableBalance: 1234,
      currentBalance: 1234,
      mask: '1234',
      name: 'Generic Checking',
      accountType: 'depository',
      accountSubType: 'checking',
    });
    const checkingAccountInactive = new BankAccount({
      bankAccountId: 'abcinactive',
      availableBalance: 1234,
      currentBalance: 1234,
      mask: '1234',
      name: 'Generic Checking',
      accountType: 'depository',
      accountSubType: 'checking',
      status: 'inactive',
    });
    const savingsAccount = new BankAccount({
      bankAccountId: 'abd',
      availableBalance: 1234,
      currentBalance: 1234,
      mask: '1234',
      name: 'Generic Savings',
      accountType: 'depository',
      accountSubType: 'savings',
    });
    const savingsAccountInactive = new BankAccount({
      bankAccountId: 'abdinactive',
      availableBalance: 1234,
      currentBalance: 1234,
      mask: '1234',
      name: 'Generic Savings',
      accountType: 'depository',
      accountSubType: 'savings',
      status: 'inactive',
    });
    const creditCard = new BankAccount({
      bankAccountId: 'abe',
      availableBalance: 1234,
      currentBalance: 1234,
      mask: '1234',
      name: 'Generic Credit Card',
      accountType: 'credit',
      accountSubType: 'credit card',
    });
    const autoInactive = new BankAccount({
      bankAccountId: 'autoinactive',
      availableBalance: 1234,
      currentBalance: 1234,
      mask: '1234',
      name: 'Generic Auto Loan',
      accountType: 'loan',
      accountSubType: 'auto',
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
