import {
  FETCH_BANK_ACCOUNT_FAILURE,
  FETCH_BANK_ACCOUNT_SUCCESS,
  FETCH_BANK_ACCOUNTS_REQUEST
} from "shared/bankAccounts/actions";
import request from "shared/util/request";
import BankAccount from "data/BankAccount";
import { Map } from 'immutable';

export const fetchBankAccountsRequest = {
  type: FETCH_BANK_ACCOUNTS_REQUEST,
};

export const fetchBankAccountsFailure = {
  type: FETCH_BANK_ACCOUNT_FAILURE,
};

export default function fetchBankAccounts() {
  return dispatch => {
    return request().get('/api/bank_accounts')
      .then(result => {
        dispatch({
          type: FETCH_BANK_ACCOUNT_SUCCESS,
          payload: Map<number, BankAccount>().withMutations(map => {
            (result.data.bank_accounts || []).forEach((bankAccount: BankAccount) => {
              map.set(bankAccount.bankAccountId, bankAccount);
            })
          }),
        });
      })
      .catch(error => {
        dispatch(fetchBankAccountsFailure);
        throw error;
      });
  }
}
