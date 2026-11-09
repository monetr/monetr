import BankAccount, { BankAccountSubType, BankAccountType } from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import Link, { LinkType } from '@monetr/interface/models/Link';
import type SimpleFinLink from '@monetr/interface/models/SimpleFinLink';
import { SimpleFinLinkStatus } from '@monetr/interface/models/SimpleFinLink';
import type User from '@monetr/interface/models/User';
import type { WithJsonValues } from '@monetr/interface/util/json';

// The strict Link constructor wants the whole JSON shape, but these tests only care about a field or two. This helper
// fills in sensible defaults so each test can just override the bits it cares about.
function fixture(overrides: Partial<WithJsonValues<Link>>): Link {
  return new Link({
    linkId: ID.from<Link>('link_test'),
    lunchFlowLinkId: null,
    simpleFinLinkId: null,
    linkType: LinkType.Plaid,
    institutionName: 'Test Institution',
    description: null,
    bankAccountOrder: [],
    updatedAt: new Date().toISOString(),
    createdAt: new Date().toISOString(),
    createdBy: ID.from<User>('user_test'),
    plaidLink: null,
    lunchFlowLink: null,
    simpleFinLink: null,
    ...overrides,
  });
}

describe('links', () => {
  it('will detect manual', () => {
    const link = fixture({
      linkType: LinkType.Manual,
    });
    expect(link.getIsManual()).toBeTruthy();
    expect(link.getIsPlaid()).toBeFalsy();
  });

  it('will detect plaid', () => {
    const link = fixture({
      linkType: LinkType.Plaid,
    });
    expect(link.getIsManual()).toBeFalsy();
  });

  it('will handle custom names', () => {
    const link = fixture({
      institutionName: 'Original',
    });

    expect(link.getName()).toBe('Original');
  });

  it('will detect simplefin', () => {
    const link = fixture({
      linkType: LinkType.SimpleFin,
      simpleFinLinkId: ID.from<SimpleFinLink>('sfx_01hy4rcmadc01d2kzv7vynbxxx'),
      simpleFinLink: {
        simpleFinLinkId: ID.from<SimpleFinLink>('sfx_01hy4rcmadc01d2kzv7vynbxxx'),
        name: 'Mercury Checking',
        apiUrl: 'https://beta-bridge.simplefin.org/simplefin',
        status: SimpleFinLinkStatus.Active,
        lastErrors: [],
        lastManualSync: null,
        lastSuccessfulUpdate: null,
        lastAttemptedUpdate: null,
        createdAt: '2026-11-09T00:00:00Z',
        createdBy: ID.from<User>('user_01hy4rcmadc01d2kzv7vynbxxx'),
        updatedAt: '2026-11-09T00:00:00Z',
        deletedAt: null,
      },
    });
    expect(link.getIsSimpleFin()).toBeTruthy();
    expect(link.getIsLunchFlow()).toBeFalsy();
    expect(link.getIsPlaid()).toBeFalsy();
    expect(link.getIsManual()).toBeFalsy();
  });

  it('will not detect simplefin without the simplefin link', () => {
    const link = fixture({
      linkType: LinkType.SimpleFin,
    });
    expect(link.getIsSimpleFin()).toBeFalsy();
  });

  it('will be in error when the simplefin link is in error', () => {
    const link = fixture({
      linkType: LinkType.SimpleFin,
      simpleFinLinkId: ID.from<SimpleFinLink>('sfx_01hy4rcmadc01d2kzv7vynbxxx'),
      simpleFinLink: {
        simpleFinLinkId: ID.from<SimpleFinLink>('sfx_01hy4rcmadc01d2kzv7vynbxxx'),
        name: 'Mercury Checking',
        apiUrl: 'https://beta-bridge.simplefin.org/simplefin',
        status: SimpleFinLinkStatus.Error,
        lastErrors: [],
        lastManualSync: null,
        lastSuccessfulUpdate: null,
        lastAttemptedUpdate: null,
        createdAt: '2026-11-09T00:00:00Z',
        createdBy: ID.from<User>('user_01hy4rcmadc01d2kzv7vynbxxx'),
        updatedAt: '2026-11-09T00:00:00Z',
        deletedAt: null,
      },
    });
    expect(link.getIsError()).toBeTruthy();
    expect(link.getIsDeactivated()).toBeFalsy();
  });

  it('will be deactivated when the simplefin link is deactivated', () => {
    const link = fixture({
      linkType: LinkType.SimpleFin,
      simpleFinLinkId: ID.from<SimpleFinLink>('sfx_01hy4rcmadc01d2kzv7vynbxxx'),
      simpleFinLink: {
        simpleFinLinkId: ID.from<SimpleFinLink>('sfx_01hy4rcmadc01d2kzv7vynbxxx'),
        name: 'Mercury Checking',
        apiUrl: 'https://beta-bridge.simplefin.org/simplefin',
        status: SimpleFinLinkStatus.Deactivated,
        lastErrors: [],
        lastManualSync: null,
        lastSuccessfulUpdate: null,
        lastAttemptedUpdate: null,
        createdAt: '2026-11-09T00:00:00Z',
        createdBy: ID.from<User>('user_01hy4rcmadc01d2kzv7vynbxxx'),
        updatedAt: '2026-11-09T00:00:00Z',
        deletedAt: null,
      },
    });
    expect(link.getIsDeactivated()).toBeTruthy();
    expect(link.getIsError()).toBeFalsy();
  });

  it('can archive a simplefin bank account', () => {
    const link = fixture({
      linkType: LinkType.SimpleFin,
      simpleFinLinkId: ID.from<SimpleFinLink>('sfx_01hy4rcmadc01d2kzv7vynbxxx'),
      simpleFinLink: {
        simpleFinLinkId: ID.from<SimpleFinLink>('sfx_01hy4rcmadc01d2kzv7vynbxxx'),
        name: 'Mercury Checking',
        apiUrl: 'https://beta-bridge.simplefin.org/simplefin',
        status: SimpleFinLinkStatus.Active,
        lastErrors: [],
        lastManualSync: null,
        lastSuccessfulUpdate: null,
        lastAttemptedUpdate: null,
        createdAt: '2026-11-09T00:00:00Z',
        createdBy: ID.from<User>('user_01hy4rcmadc01d2kzv7vynbxxx'),
        updatedAt: '2026-11-09T00:00:00Z',
        deletedAt: null,
      },
    });
    // Unlike Plaid this doesn't need the account to be inactive first
    const bankAccount = new BankAccount({
      bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
      linkId: ID.from<Link>('link_test'),
      lunchFlowBankAccountId: null,
      simpleFinBankAccountId: null,
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
      lastUpdated: '2026-11-09T00:00:00Z',
      createdAt: '2026-11-09T00:00:00Z',
      createdBy: 'user_01hy4rcmadc01d2kzv7vynbxxx',
      deletedAt: null,
      plaidBankAccount: null,
      lunchFlowBankAccount: null,
      simpleFinBankAccount: null,
    });
    expect(link.getCanArchiveBankAccount(bankAccount)).toBeTruthy();
  });
});
