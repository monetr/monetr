export enum LinkType {
  Unknown = 0,
  Plaid = 1,
  Manual = 2,
}

export enum LinkStatus {
  Unknown = 0,
  Pending = 1,
  Setup = 2,
  Error = 3,
}

export default class Link {
  linkId: number;
  linkType: LinkType;
  linkStatus: LinkStatus;
  errorCode: string | null;
  institutionId: number;
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

  getIsPlaid(): boolean {
    return this.linkType === LinkType.Plaid;
  }
}
