import { ID, idPrefix } from '@monetr/interface/models/ID';
import LunchFlowLink, { LunchFlowLinkStatus } from '@monetr/interface/models/LunchFlowLink';
import PlaidLink, { PlaidLinkStatus } from '@monetr/interface/models/PlaidLink';
import User from '@monetr/interface/models/User';
import { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export enum LinkType {
  Unknown = 'unknown',
  Plaid = 'plaid',
  Manual = 'manual',
  Stripe = 'stripe',
  LunchFlow = 'lunch_flow',
}

export const errorMessages: Partial<Record<string, string>> = {
  ITEM_LOGIN_REQUIRED: "This link's authentication has expired and needs to be re-authenticated.",
};

/**
 * A Link is used to represent an institution and it's bank accounts. An account can have multiple links and can have
 * multiple links for the same institution. But typically links will represent a unique group of bank accounts for a
 * an institution. A group of bank accounts within a single "login" for that institution.
 */
export default class Link {
  readonly [idPrefix] = 'link';

  /**
   * Represents the global unique identifier for a group of bank accounts in monetr.
   * This value is generated automatically by the API upon creation, and cannot be changed.
   */
  linkId: ID<Link>;
  lunchFlowLinkId: ID<LunchFlowLink> | null;
  linkType: LinkType;
  institutionName: string;
  description: string | null;
  updatedAt: Date;
  createdAt: Date;
  createdBy: ID<User>;

  plaidLink: PlaidLink | null;
  lunchFlowLink: LunchFlowLink | null;

  constructor(data: WithJsonValues<Link>) {}

  getName(): string {
    return this.institutionName;
  }

  getIsManual(): boolean {
    return this.linkType === LinkType.Manual;
  }

  getIsPlaid(): boolean {
    return this.linkType === LinkType.Plaid && Boolean(this.plaidLink);
  }

  getCanUpdateAccountSelection(): boolean {
    return this.getIsPlaid() && this.plaidLink?.newAccountsAvailable === true;
  }

  getIsError(): boolean {
    return this.plaidLink?.status === PlaidLinkStatus.Error || this.lunchFlowLink?.status === LunchFlowLinkStatus.Error;
  }

  getIsPendingExpiration(): boolean {
    return this.plaidLink?.status === PlaidLinkStatus.PendingExpiration;
  }

  getIsRevoked(): boolean {
    return this.plaidLink?.status === PlaidLinkStatus.Revoked;
  }

  getErrorMessage(): string | null {
    const code = this.plaidLink?.status;
    if (!code) {
      return null;
    }
    return errorMessages[code] ?? null;
  }
}
