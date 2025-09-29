
import PlaidLink, { PlaidLinkStatus } from '@monetr/interface/models/PlaidLink';
import parseDate from '@monetr/interface/util/parseDate';

export enum LinkType {
  Unknown = 0,
  Plaid = 1,
  Manual = 2,
}

export const errorMessages = {
  'ITEM_LOGIN_REQUIRED': 'This link\'s authentication has expired and needs to be re-authenticated.',
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
  linkId: string;
  linkType: LinkType;
  institutionName: string;
  description: string | null;
  updatedAt: Date;
  createdAt: Date;
  createdBy: string;

  plaidLink: PlaidLink | null;

  constructor(data?: Partial<Link>) {
    if (data) Object.assign(this, {
      ...data,
      plaidLink: data?.plaidLink && new PlaidLink(data.plaidLink),
      updatedAt: parseDate(data?.updatedAt),
      createdAt: parseDate(data?.createdAt),
    });
  }

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
    return this.plaidLink?.status === PlaidLinkStatus.Error;
  }

  getIsPendingExpiration(): boolean {
    return this.plaidLink?.status === PlaidLinkStatus.PendingExpiration;
  }

  getIsRevoked(): boolean {
    return this.plaidLink?.status === PlaidLinkStatus.Revoked;
  }

  getErrorMessage(): string | null {
    const code = this.plaidLink?.status;
    return errorMessages[code] || null;
  }
}
