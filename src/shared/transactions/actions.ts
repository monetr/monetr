import Transaction from "data/Transaction";
import { Logout } from "shared/authentication/actions";

export const FETCH_TRANSACTIONS_REQUEST = 'FETCH_TRANSACTIONS_REQUEST';
export const FETCH_TRANSACTIONS_FAILURE = 'FETCH_TRANSACTIONS_FAILURE';
export const FETCH_TRANSACTIONS_SUCCESS = 'FETCH_TRANSACTIONS_SUCCESS';

export interface FetchTransactionsSuccess {
  type: typeof FETCH_TRANSACTIONS_SUCCESS;
  bankAccountId: number;
  payload: Transaction[];
}

export interface FetchTransactionsRequest {
  type: typeof FETCH_TRANSACTIONS_REQUEST;
}

export interface FetchTransactionsFailure {
  type: typeof FETCH_TRANSACTIONS_FAILURE;
}

export type TransactionActions = FetchTransactionsSuccess | FetchTransactionsRequest | FetchTransactionsFailure | Logout
