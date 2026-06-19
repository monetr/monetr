import { ID } from '@monetr/interface/models/ID';
import Link, { LinkType } from '@monetr/interface/models/Link';
import type User from '@monetr/interface/models/User';
import type { WithJsonValues } from '@monetr/interface/util/json';

// The strict Link constructor wants the whole JSON shape, but these tests only care about a field or two. This helper
// fills in sensible defaults so each test can just override the bits it cares about.
function fixture(overrides: Partial<WithJsonValues<Link>>): Link {
  return new Link({
    linkId: ID.from<Link>('link_test'),
    lunchFlowLinkId: null,
    linkType: LinkType.Plaid,
    institutionName: 'Test Institution',
    description: null,
    updatedAt: new Date().toISOString(),
    createdAt: new Date().toISOString(),
    createdBy: ID.from<User>('user_test'),
    plaidLink: null,
    lunchFlowLink: null,
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
});
