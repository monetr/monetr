import Transaction from "data/Transaction";
import { Logout } from "shared/authentication/actions";
import { ChangeBankAccount } from "shared/bankAccounts/actions";

export const FETCH_TRANSACTIONS_REQUEST = 'FETCH_TRANSACTIONS_REQUEST';
export const FETCH_TRANSACTIONS_FAILURE = 'FETCH_TRANSACTIONS_FAILURE';
export const FETCH_TRANSACTIONS_SUCCESS = 'FETCH_TRANSACTIONS_SUCCESS';
export const CHANGE_SELECTED_TRANSACTION = 'CHANGE_SELECTED_TRANSACTION';

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

export interface ChangeSelectedTransaction {
  type: typeof CHANGE_SELECTED_TRANSACTION;
  transactionId: number;
}

export type TransactionActions =
  FetchTransactionsSuccess
  | FetchTransactionsRequest
  | FetchTransactionsFailure
  | Logout
  | ChangeBankAccount
  | ChangeSelectedTransaction
