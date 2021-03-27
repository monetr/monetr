import Spending from "data/Spending";
import { Map } from 'immutable';
import { Logout } from "shared/authentication/actions";
import { ChangeBankAccount } from "shared/bankAccounts/actions";

export const FETCH_SPENDING_REQUEST = 'FETCH_SPENDING_REQUEST';
export const FETCH_SPENDING_FAILURE = 'FETCH_SPENDING_FAILURE';
export const FETCH_SPENDING_SUCCESS = 'FETCH_SPENDING_SUCCESS';

export interface FetchSpendingSuccess {
  type: typeof FETCH_SPENDING_SUCCESS;
  payload: Map<number, Map<number, Spending>>;
}

export interface FetchSpendingFailure {
  type: typeof FETCH_SPENDING_FAILURE;
}

export interface FetchSpendingRequest {
  type: typeof FETCH_SPENDING_REQUEST;
}


export type SpendingActions =
  FetchSpendingRequest
  | FetchSpendingFailure
  | FetchSpendingSuccess
  | Logout
  | ChangeBankAccount
