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

export default class Link {
  linkId: number;
  linkType: LinkType;
  linkStatus: LinkStatus;
  errorCode: string | null;
  institutionId: number;
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
