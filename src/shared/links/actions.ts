import Link from "data/Link";
import { Logout } from "shared/authentication/actions";
import { Map } from 'immutable';

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

export type LinkActions =
  FetchLinksSuccess
  | FetchLinksRequest
  | FetchLinksFailure
  | CreateLinksRequest
  | CreateLinksFailure
  | CreateLinksSuccess
  | Logout
