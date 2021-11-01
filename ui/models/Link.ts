import { Moment } from "moment";
import { parseToMomentMaybe } from "util/parseToMoment";

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

export const errorMessages = {
  'ITEM_LOGIN_REQUIRED': `This link's authentication has expired and needs to be re-authenticated.`
};

/**
  * A Link is used to represent an institution and it's bank accounts. An account can have multiple links and can have
  * multiple links for the same institution. But typically links will represent a unique group of bank accounts for a
  * an institution. A group of bank accounts within a single "login" for that institution.
  */
export default class Link {
  /**
   * Represents the global unique identifier for a group of bank accounts in monetr.
   * This value is generated automatically by the API upon creation, and cannot be changed.
   */
  linkId: number;
  linkType: LinkType;
  linkStatus: LinkStatus;
  errorCode: string | null;
  plaidInstitutionId: string | null;
  institutionName: string;
  customInstitutionName?: string;
  createdByUserId: number;
  lastSuccessfulUpdate: Moment | null;

  constructor(data: Partial<Link>) {
    if (data) {
      Object.assign(this, {
        ...data,
        lastSuccessfulUpdate: parseToMomentMaybe(data.lastSuccessfulUpdate),
      });
    }
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

  getIsError(): boolean {
    return this.linkStatus === LinkStatus.Error || this.errorCode != null;
  }

  getErrorMessage(): string | null {
    return errorMessages[this.errorCode] || null;
  }
}
