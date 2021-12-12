import Transaction from 'models/Transaction';
import { useDispatch, useSelector, useStore } from 'react-redux';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import {
  FETCH_TRANSACTIONS_FAILURE,
  FETCH_TRANSACTIONS_REQUEST,
  FETCH_TRANSACTIONS_SUCCESS
} from 'shared/transactions/actions';
import { getHasAnyTransactions } from 'shared/transactions/selectors/getHasAnyTransactions';
import request from 'shared/util/request';

function useFetchInitialTransactionsIfNeeded(): () => Promise<void> {
  // This will prevent the fetchInitialTransactions function returned from changing as the state changes. This will make
  // it so we can evaluate the state and selectors we need at call-time, not at the time this hook is created.
  const { dispatch, getState } = useStore();

  return (): Promise<void> => {
    const selectedBankAccountId = getSelectedBankAccountId(getState());
    const hasAnyTransactions = getHasAnyTransactions(getState());

    if (!selectedBankAccountId || hasAnyTransactions) {
      console.debug('skipping retrieving transactions due to:', {
        hasSelectedBankAccount: !!selectedBankAccountId,
        hasAnyTransactions,
      });
      // If the user does not have a bank account selected, then there are no transactions we can request.
      return Promise.resolve();
    }

    dispatch({
      type: FETCH_TRANSACTIONS_REQUEST,
    });

    return request()
      .get(`/bank_accounts/${ selectedBankAccountId }/transactions`)
      .then(result => void dispatch({
        type: FETCH_TRANSACTIONS_SUCCESS,
        bankAccountId: selectedBankAccountId,
        payload: result.data.map(item => new Transaction(item)),
      }))
      .catch(error => {
        dispatch({
          type: FETCH_TRANSACTIONS_FAILURE,
        });
        throw error;
      });
  }
}

export default useFetchInitialTransactionsIfNeeded;
