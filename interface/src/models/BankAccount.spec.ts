import BankAccount, { BankAccountSubType, BankAccountType } from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type Link from '@monetr/interface/models/Link';

// If some of these tests seem goofy. I'm using them as a way to play around with how typescript constructors work.
describe('bank accounts', () => {
  it('will construct with fields', () => {
    const now = new Date();
    const account = new BankAccount({
      bankAccountId: ID.from<BankAccount>('bac_test'),
      linkId: ID.from<Link>('link_test'),
      lunchFlowBankAccountId: null,
      mask: '1234',
      name: 'Generic Checking',
      originalName: 'Generic Checking',
      status: 'active',
      accountType: BankAccountType.Depository,
      accountSubType: BankAccountSubType.Checking,
      currency: 'USD',
      currentBalance: 0,
      availableBalance: 0,
      limitBalance: null,
      // WithJsonValues lets us hand the constructor the raw string form of a date and trust it to parse it back into an
      // actual Date for us, which is exactly what the API does.
      lastUpdated: now.toISOString(),
      createdAt: now.toISOString(),
      createdBy: 'user_test',
      deletedAt: null,
      plaidBankAccount: null,
      lunchFlowBankAccount: null,
    });
    expect(account.lastUpdated).toEqual(now);
  });
});
