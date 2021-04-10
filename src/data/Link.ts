export enum LinkType {
  Unknown = 0,
  Plaid = 1,
  Manual = 2,
}

export default class Link {
  linkId: number;
  linkType: LinkType;
  institutionId: string;
  institutionName: string;
  customInstitutionName?: string;
  createdByUserId: number;

  constructor(data: Partial<Link>) {
    Object.assign(this, data)
  }

  getName(): string {
    return this.customInstitutionName ?? this.institutionName;
  }

  getIsManual(): boolean {
    return this.linkType === LinkType.Manual;
  }
}
