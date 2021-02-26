export interface LinkFields {
  linkId: number;
  institutionId: string;
  institutionName: string;
  customInstitutionName?: string;
  createdByUserId: number;
}

export default class Link implements LinkFields {
  linkId: number;
  institutionId: string;
  institutionName: string;
  customInstitutionName?: string;
  createdByUserId: number;

  constructor(data: LinkFields) {
    Object.assign(this, data)
  }

  public getName(): string {
    return this.customInstitutionName ?? this.institutionName;
  }
}
