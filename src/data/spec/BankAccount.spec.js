import BankAccount from "data/BankAccount";


describe('bank accounts', () => {
  it('will return available balance', () => {
    const account = new BankAccount();
    account.availableBalance = 1000;

    expect(account.getAvailableBalanceString()).toBe('$10.00');
  });
});
