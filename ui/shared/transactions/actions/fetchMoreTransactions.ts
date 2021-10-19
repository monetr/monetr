import Transaction from 'data/Transaction';
import { Dispatch } from 'redux';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import {
  FETCH_TRANSACTIONS_FAILURE,
  FETCH_TRANSACTIONS_REQUEST,
  FETCH_TRANSACTIONS_SUCCESS
} from 'shared/transactions/actions';
import { getTransactions } from 'shared/transactions/selectors/getTransactions';
import request from 'shared/util/request';

interface GetState {
  (): object
}

interface ActionWithState {
  (dispatch: Dispatch, getState: GetState): Promise<void>
}

export default function fetchMoreTransactions(): ActionWithState {
  return (dispatch, getState) => {
    const selectedBankAccountId = getSelectedBankAccountId(getState());
    if (!selectedBankAccountId) {
      // If the user does not have a bank account selected, then there are no transactions we can request.
      return Promise.resolve();
    }

    const offset = getTransactions(getState())?.count() || 0;
    const limit = 25;

    dispatch({
      type: FETCH_TRANSACTIONS_REQUEST,
    });

    return request().get(`/bank_accounts/${ selectedBankAccountId }/transactions?offset=${ offset }&limit=${ limit }`)
      .then(result => {
        dispatch({
          type: FETCH_TRANSACTIONS_SUCCESS,
          bankAccountId: selectedBankAccountId,
          payload: result.data.map(item => new Transaction(item)),
        });
      })
      .catch(error => {
        dispatch({
          type: FETCH_TRANSACTIONS_FAILURE,
        });
        throw error;
      });
  };
}