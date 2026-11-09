import { waitFor } from '@testing-library/react';

import SimpleFinSetupSyncItem from '@monetr/interface/components/setup/simplefin/SimpleFinSetupSyncItem';
import BankAccount, { BankAccountSubType, BankAccountType } from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type Link from '@monetr/interface/models/Link';
import type SimpleFinBankAccount from '@monetr/interface/models/SimpleFinBankAccount';
import { SimpleFinBankAccountStatus } from '@monetr/interface/models/SimpleFinBankAccount';
import type SimpleFinLink from '@monetr/interface/models/SimpleFinLink';
import type User from '@monetr/interface/models/User';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('simplefin setup sync item', () => {
  it('will show a spinner while waiting', async () => {
    const bankAccount = new BankAccount({
      bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
      linkId: ID.from<Link>('link_01hy4rcmadc01d2kzv7vynbxxx'),
      lunchFlowBankAccountId: null,
      simpleFinBankAccountId: ID.from<SimpleFinBankAccount>('sbac_01hy4rbb1gjdek7h2xmgy5pnwk'),
      mask: null,
      name: 'Mercury Checking',
      originalName: 'Mercury Checking',
      status: 'active',
      accountType: BankAccountType.Depository,
      accountSubType: BankAccountSubType.Checking,
      currency: 'USD',
      currentBalance: 12345,
      availableBalance: 12345,
      limitBalance: null,
      lastUpdated: '2026-03-16T06:00:00Z',
      createdAt: '2026-03-15T06:00:00Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      deletedAt: null,
      plaidBankAccount: null,
      lunchFlowBankAccount: null,
      simpleFinBankAccount: {
        simpleFinBankAccountId: ID.from<SimpleFinBankAccount>('sbac_01hy4rbb1gjdek7h2xmgy5pnwk'),
        simpleFinLinkId: ID.from<SimpleFinLink>('sfx_01hy4rbb1gjdek7h2xmgy5pnwk'),
        simpleFinId: 'ACT-01hy4rbb1gjdek7h2xmgy5pnwk',
        connectionId: 'CON-01hy4rcmadc01d2kzv7vynbxxx',
        name: 'Mercury Checking',
        institutionName: 'Mercury',
        currency: 'USD',
        status: SimpleFinBankAccountStatus.Active,
        currentBalance: 12345,
        availableBalance: 12345,
        createdAt: '2026-03-15T06:00:00Z',
        createdBy: ID.from<User>('user_01hym36e8ewaq0hxssb1m3k4ha'),
        updatedAt: '2026-03-16T06:00:00Z',
        deletedAt: null,
      },
    });

    const world = testRenderer(<SimpleFinSetupSyncItem bankAccount={bankAccount} />, {
      initialRoute: '/link/create/simplefin/link_01hy4rcmadc01d2kzv7vynbxxx/sync',
    });

    await waitFor(() => expect(world.getByTestId('simplefin-sync-loading')).toBeVisible());
    expect(world.getByText('Waiting')).toBeVisible();
    expect(world.getByText('Mercury')).toBeVisible();
    expect(world.queryByTestId('simplefin-sync-complete')).toBeNull();
    expect(world.queryByTestId('simplefin-sync-error')).toBeNull();
  });

  it('will show a check when complete and no spinner', async () => {
    const bankAccount = new BankAccount({
      bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
      linkId: ID.from<Link>('link_01hy4rcmadc01d2kzv7vynbxxx'),
      lunchFlowBankAccountId: null,
      simpleFinBankAccountId: ID.from<SimpleFinBankAccount>('sbac_01hy4rbb1gjdek7h2xmgy5pnwk'),
      mask: null,
      name: 'Mercury Checking',
      originalName: 'Mercury Checking',
      status: 'active',
      accountType: BankAccountType.Depository,
      accountSubType: BankAccountSubType.Checking,
      currency: 'USD',
      currentBalance: 12345,
      availableBalance: 12345,
      limitBalance: null,
      lastUpdated: '2026-03-16T06:00:00Z',
      createdAt: '2026-03-15T06:00:00Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      deletedAt: null,
      plaidBankAccount: null,
      lunchFlowBankAccount: null,
      simpleFinBankAccount: null,
    });

    const world = testRenderer(<SimpleFinSetupSyncItem bankAccount={bankAccount} status='complete' />, {
      initialRoute: '/link/create/simplefin/link_01hy4rcmadc01d2kzv7vynbxxx/sync',
    });

    await waitFor(() => expect(world.getByTestId('simplefin-sync-complete')).toBeVisible());
    expect(world.queryByTestId('simplefin-sync-loading')).toBeNull();
    expect(world.queryByTestId('simplefin-sync-error')).toBeNull();
  });

  it('will show an x on error and no spinner', async () => {
    const bankAccount = new BankAccount({
      bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
      linkId: ID.from<Link>('link_01hy4rcmadc01d2kzv7vynbxxx'),
      lunchFlowBankAccountId: null,
      simpleFinBankAccountId: ID.from<SimpleFinBankAccount>('sbac_01hy4rbb1gjdek7h2xmgy5pnwk'),
      mask: null,
      name: 'Mercury Checking',
      originalName: 'Mercury Checking',
      status: 'active',
      accountType: BankAccountType.Depository,
      accountSubType: BankAccountSubType.Checking,
      currency: 'USD',
      currentBalance: 12345,
      availableBalance: 12345,
      limitBalance: null,
      lastUpdated: '2026-03-16T06:00:00Z',
      createdAt: '2026-03-15T06:00:00Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      deletedAt: null,
      plaidBankAccount: null,
      lunchFlowBankAccount: null,
      simpleFinBankAccount: null,
    });

    const world = testRenderer(<SimpleFinSetupSyncItem bankAccount={bankAccount} status='error' />, {
      initialRoute: '/link/create/simplefin/link_01hy4rcmadc01d2kzv7vynbxxx/sync',
    });

    await waitFor(() => expect(world.getByTestId('simplefin-sync-error')).toBeVisible());
    expect(world.queryByTestId('simplefin-sync-loading')).toBeNull();
    expect(world.queryByTestId('simplefin-sync-complete')).toBeNull();
  });
});
