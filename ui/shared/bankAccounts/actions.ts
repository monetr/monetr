import { Map } from 'immutable';
import BankAccount from 'models/BankAccount';
import { LogoutActions } from 'shared/authentication/actions';
import { RemoveLinkSuccess } from 'shared/links/actions';

export const FETCH_BANK_ACCOUNTS_REQUEST = 'FETCH_BANK_ACCOUNTS_REQUEST';
export const FETCH_BANK_ACCOUNTS_FAILURE = 'FETCH_BANK_ACCOUNT_FAILURE';
export const FETCH_BANK_ACCOUNTS_SUCCESS = 'FETCH_BANK_ACCOUNT_SUCCESS';

export const CHANGE_BANK_ACCOUNT = 'CHANGE_BANK_ACCOUNT';

export interface ChangeBankAccount {
  type: typeof CHANGE_BANK_ACCOUNT;
  payload: number;
}

export interface FetchBankAccountsSuccess {
  type: typeof FETCH_BANK_ACCOUNTS_SUCCESS,
  payload: Map<number, BankAccount>;
}

export interface FetchBankAccountsRequest {
  type: typeof FETCH_BANK_ACCOUNTS_REQUEST,
}

export interface FetchBankAccountsFailure {
  type: typeof FETCH_BANK_ACCOUNTS_FAILURE,
}

export type BankAccountActions =
  ChangeBankAccount
  | FetchBankAccountsSuccess
  | FetchBankAccountsRequest
  | FetchBankAccountsFailure
  | RemoveLinkSuccess
  | LogoutActions
