import Transaction from "data/Transaction";
import { getSelectedBankAccountId } from "shared/bankAccounts/selectors/getSelectedBankAccountId";
import {
  FETCH_TRANSACTIONS_FAILURE,
  FETCH_TRANSACTIONS_REQUEST,
  FETCH_TRANSACTIONS_SUCCESS
} from "shared/transactions/actions";
import { getHasAnyTransactions } from "shared/transactions/selectors/getHasAnyTransactions";
import request from "shared/util/request";
import { AppDispatch, AppState } from 'store';

interface ActionWithState {
  (dispatch: AppDispatch, getState: () => AppState): Promise<void>
}

export default function fetchInitialTransactionsIfNeeded(): ActionWithState {
  return (dispatch, getState) => {
    const selectedBankAccountId = getSelectedBankAccountId(getState());
    if (!selectedBankAccountId) {
      // If the user does not have a bank account selected, then there are no transactions we can request.
      return Promise.resolve();
    }

    const hasAnyTransactions = getHasAnyTransactions(getState());
    if (hasAnyTransactions) {
      return Promise.resolve();
    }

    dispatch({
      type: FETCH_TRANSACTIONS_REQUEST,
    });

    return request().get(`/bank_accounts/${ selectedBankAccountId }/transactions`)
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
      })
  };
}
