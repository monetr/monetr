import request from "shared/util/request";
import { getSelectedBankAccountId } from "shared/bankAccounts/selectors/getSelectedBankAccountId";
import {
  FETCH_TRANSACTIONS_FAILURE,
  FETCH_TRANSACTIONS_REQUEST,
  FETCH_TRANSACTIONS_SUCCESS
} from "shared/transactions/actions";
import Transaction from "data/Transaction";

interface Dispatch {
  (action: {}): void
}

interface GetState {
  (): object
}

interface ActionWithState {
  (dispatch: Dispatch, getState: GetState): Promise<void>
}

export default function fetchInitialTransactionsIfNeeded(): ActionWithState {
  return (dispatch, getState) => {
    const selectedBankAccountId = getSelectedBankAccountId(getState());
    if (!selectedBankAccountId) {
      // If the user does not have a bank account selected, then there are no transactions we can request.
      return Promise.resolve();
    }

    // TODO Check and see if there are any transactions for the selected bank account first. If there are do nothing,
    //  if there are not then request the first page of transactions.

    dispatch({
      type: FETCH_TRANSACTIONS_REQUEST,
    });

    return request().get(`/api/bank_accounts/${ selectedBankAccountId }/transactions`)
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
