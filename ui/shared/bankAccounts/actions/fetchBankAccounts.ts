import { Map } from 'immutable';
import BankAccount from 'models/BankAccount';
import {
  FETCH_BANK_ACCOUNTS_FAILURE,
  FETCH_BANK_ACCOUNTS_REQUEST,
  FETCH_BANK_ACCOUNTS_SUCCESS,
} from 'shared/bankAccounts/actions';
import request from 'shared/util/request';

export const fetchBankAccountsRequest = {
  type: FETCH_BANK_ACCOUNTS_REQUEST,
};

export const fetchBankAccountsFailure = {
  type: FETCH_BANK_ACCOUNTS_FAILURE,
};

export default function fetchBankAccounts() {
  return dispatch => {
    dispatch(fetchBankAccountsRequest);

    return request().get('/bank_accounts')
      .then(result => {
        dispatch({
          type: FETCH_BANK_ACCOUNTS_SUCCESS,
          payload: Map<number, BankAccount>().withMutations(map => {
            (result.data || []).forEach((bankAccount: BankAccount) => {
              map.set(bankAccount.bankAccountId, new BankAccount(bankAccount));
            });
          }),
        });
      })
      .catch(error => {
        dispatch(fetchBankAccountsFailure);
        throw error;
      });
  };
}
