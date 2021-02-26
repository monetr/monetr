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

export type LinkActions = FetchLinksSuccess | FetchLinksRequest | FetchLinksFailure | Logout
