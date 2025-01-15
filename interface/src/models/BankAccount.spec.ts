import BankAccount from '@monetr/interface/models/BankAccount';

// If some of these tests seem goofy. I'm using them as a way to play around with how typescript constructors work.
describe('bank accounts', () => {
  it('will construct with fields', () => {
    const now = new Date();
    const account = new BankAccount({
      lastUpdated: now.toISOString() as any,
    });
    expect(account.lastUpdated).toEqual(now);
  });
});
