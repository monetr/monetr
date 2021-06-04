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
}
