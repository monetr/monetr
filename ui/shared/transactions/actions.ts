import Spending from 'models/Spending';
import Transaction from "models/Transaction";
import { LogoutActions } from 'shared/authentication/actions';
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

export enum UpdateTransaction {
  Request = 'UpdateTransactionRequest',
  Failure = 'UpdateTransactionFailure',
  Success = 'UpdateTransactionSuccess',
}

export interface UpdateTransactionRequest {
  type: typeof UpdateTransaction.Request;
}

export interface UpdateTransactionFailure {
  type: typeof UpdateTransaction.Failure;
}

export interface UpdateTransactionSuccess {
  type: typeof UpdateTransaction.Success;
  payload: {
    transaction: Transaction;
    spending: Spending[];
  };
}

export type TransactionActions =
  FetchTransactionsSuccess
  | FetchTransactionsRequest
  | FetchTransactionsFailure
  | LogoutActions
  | ChangeBankAccount
  | ChangeSelectedTransaction
  | UpdateTransactionRequest
  | UpdateTransactionFailure
  | UpdateTransactionSuccess
