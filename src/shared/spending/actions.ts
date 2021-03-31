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

export enum CreateSpending {
  Request = 'CreateSpendingRequest',
  Failure = 'CreateSpendingFailure',
  Success = 'CreateSpendingSuccess',
}

export interface CreateSpendingRequest {
  type: typeof CreateSpending.Request;
}

export interface CreateSpendingFailure {
  type: typeof CreateSpending.Failure;
}

export interface CreateSpendingSuccess {
  type: typeof CreateSpending.Success;
  payload: Spending;
}

export const SelectExpense = 'SelectExpense';

export interface SelectExpense {
  type: typeof SelectExpense;
  expenseId: number | null;
}


export type SpendingActions =
  FetchSpendingRequest
  | FetchSpendingFailure
  | FetchSpendingSuccess
  | CreateSpendingRequest
  | CreateSpendingFailure
  | CreateSpendingSuccess
  | SelectExpense
  | Logout
  | ChangeBankAccount
