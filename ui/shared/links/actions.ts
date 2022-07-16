import { Map } from 'immutable';
import BankAccount from 'models/BankAccount';
import Link from 'models/Link';
import { LogoutActions } from 'shared/authentication/actions';

export const FETCH_LINKS_REQUEST = 'FETCH_LINKS_REQUEST';
export const FETCH_LINKS_SUCCESS = 'FETCH_LINKS_SUCCESS';
export const FETCH_LINKS_FAILURE = 'FETCH_LINKS_FAILURE';

export interface FetchLinksSuccess {
  type: typeof FETCH_LINKS_SUCCESS;
  payload: Map<number, Link>;
}

export interface FetchLinksRequest {
  type: typeof FETCH_LINKS_REQUEST;
}

export interface FetchLinksFailure {
  type: typeof FETCH_LINKS_FAILURE;
}

export enum CreateLinks {
  Request = 'CreateLinksRequest',
  Failure = 'CreateLinksFailure',
  Success = 'CreateLinksSuccess',
}

export interface CreateLinksRequest {
  type: typeof CreateLinks.Request;
}

export interface CreateLinksFailure {
  type: typeof CreateLinks.Failure;
}

export interface CreateLinksSuccess {
  type: typeof CreateLinks.Success;
  payload: Link;
}

export enum RemoveLink {
  Request = 'RemoveLinkRequest',
  Failure = 'RemoveLinkFailure',
  Success = 'RemoveLinkSuccess',
}

export interface RemoveLinkRequest {
  type: typeof RemoveLink.Request;
}

export interface RemoveLinkFailure {
  type: typeof RemoveLink.Failure;
}

export interface RemoveLinkSuccess {
  type: typeof RemoveLink.Success;
  payload: {
    link: Link;
    bankAccounts: Map<number, BankAccount>;
  };
}

export type LinkActions =
  FetchLinksSuccess
  | FetchLinksRequest
  | FetchLinksFailure
  | CreateLinksRequest
  | CreateLinksFailure
  | CreateLinksSuccess
  | RemoveLinkRequest
  | RemoveLinkFailure
  | RemoveLinkSuccess
  | LogoutActions
