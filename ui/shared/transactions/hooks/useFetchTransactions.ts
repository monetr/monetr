import Transaction from 'models/Transaction';
import { useStore } from 'react-redux';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import {
  FETCH_TRANSACTIONS_FAILURE,
  FETCH_TRANSACTIONS_REQUEST,
  FETCH_TRANSACTIONS_SUCCESS
} from 'shared/transactions/actions';
import request from 'shared/util/request';

export default function useFetchTransactions(): (offset?: number) => Promise<void> {
  const { dispatch, getState} = useStore();

  return function (offset?: number): Promise<void> {
    const selectedBankAccountId = getSelectedBankAccountId(getState());
    if (!selectedBankAccountId) {
      return Promise.resolve();
    }

    dispatch({
      type: FETCH_TRANSACTIONS_REQUEST,
    });

    return request()
      .get(`/bank_accounts/${ selectedBankAccountId }/transactions`, {
        params: offset && {
          offset: offset!
        },
      })
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
