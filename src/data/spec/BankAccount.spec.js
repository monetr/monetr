import BankAccount from "data/BankAccount";

// If some of these tests seem goofy. I'm using them as a way to play around with how typescript constructors work.
describe('bank accounts', () => {
  it('will return available balance', () => {
    const account = new BankAccount();
    account.availableBalance = 1000;

    expect(account.getAvailableBalanceString()).toBe('$10.00');
  });

  it('will construct with fields', () => {
    const account = new BankAccount({
      availableBalance: 1000,
    });

    expect(account.getAvailableBalanceString()).toBe('$10.00');
  });
});
