import Link, { LinkType } from '@monetr/interface/models/Link';

import { describe, expect, it } from 'bun:test';

describe('links', () => {
  it('will detect manual', () => {
    const link = new Link({
      linkType: LinkType.Manual,
    });
    expect(link.getIsManual()).toBeTruthy();

    link.linkType = LinkType.Plaid;
    expect(link.getIsManual()).toBeFalsy();
  });

  it('will handle custom names', () => {
    const link = new Link({
      institutionName: 'Original',
    });

    expect(link.getName()).toBe('Original');
  });
});
