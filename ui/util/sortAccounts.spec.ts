import sortAccounts from './sortAccounts';

import BankAccount from 'models/BankAccount';

describe('sort accounts', () => {
  it('will make sure that checking is the highest priority', () => {
    const checkingAccount = new BankAccount({
      'bankAccountId': 2,
      'linkId': 1,
      'availableBalance': 1234,
      'currentBalance': 1234,
      'mask': '1234',
      'name': 'Generic Checking',
      'accountType': 'depository',
      'accountSubType': 'checking',
    });
    const savingsAccount = new BankAccount({
      'bankAccountId': 1,
      'linkId': 1,
      'availableBalance': 1234,
      'currentBalance': 1234,
      'mask': '1234',
      'name': 'Generic Savings',
      'accountType': 'depository',
      'accountSubType': 'savings',
    });
    const accounts = [
      savingsAccount,
      checkingAccount,
    ];

    const result = sortAccounts(accounts);
    expect(result[0]).toBe(checkingAccount);
    expect(result[1]).toBe(savingsAccount);
  });

  it('will also handle credit card', () => {
    const checkingAccount = new BankAccount({
      'bankAccountId': 2,
      'linkId': 1,
      'availableBalance': 1234,
      'currentBalance': 1234,
      'mask': '1234',
      'name': 'Generic Checking',
      'accountType': 'depository',
      'accountSubType': 'checking',
    });
    const savingsAccount = new BankAccount({
      'bankAccountId': 1,
      'linkId': 1,
      'availableBalance': 1234,
      'currentBalance': 1234,
      'mask': '1234',
      'name': 'Generic Savings',
      'accountType': 'depository',
      'accountSubType': 'savings',
    });
    const creditCard = new BankAccount({
      'bankAccountId': 3,
      'linkId': 1,
      'availableBalance': 1234,
      'currentBalance': 1234,
      'mask': '1234',
      'name': 'Generic Credit Card',
      'accountType': 'credit',
      'accountSubType': 'credit card',
    });
    const accounts = [
      savingsAccount,
      creditCard,
      checkingAccount,
    ];

    const result = sortAccounts(accounts);
    expect(result[0]).toBe(checkingAccount);
    expect(result[1]).toBe(savingsAccount);
    expect(result[2]).toBe(creditCard);
  });
});
