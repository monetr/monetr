export default class Link {
  linkId: number;
  institutionId: string;
  institutionName: string;
  customInstitutionName?: string;
  createdByUserId: number;

  getName() {
    return this.customInstitutionName || this.institutionName;
  }
}
