import Link, { LinkType } from "models/Link";


describe('links', () => {
  it('will detect manual', () => {
    const link = new Link({
      linkType: LinkType.Manual,
    })
    expect(link.getIsManual()).toBeTruthy();

    link.linkType = LinkType.Plaid;
    expect(link.getIsManual()).toBeFalsy();
  });

  it('will handle custom names', () => {
    const link = new Link({
      institutionName: 'Original'
    });

    expect(link.getName()).toBe('Original');

    link.customInstitutionName = 'New Name';

    // Now that a custom institution name is present, this should change.
    expect(link.getName()).toBe('New Name');
  });
});
