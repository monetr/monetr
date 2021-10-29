import Balance from 'models/Balance';
import { Dispatch } from 'redux';
import { FetchBalances } from 'shared/balances/actions';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import request from 'shared/util/request';

const fetchBalancesRequest = {
  type: FetchBalances.Request,
};

const fetchBalancesFailure = {
  type: FetchBalances.Failure,
};

export default function fetchBalancesForBankAccount(bankAccountId: number|null) {
  return (dispatch: Dispatch, getState) => {
    if (!bankAccountId) {
      return Promise.resolve();
    }

    dispatch(fetchBalancesRequest);

    return request().get(`/bank_accounts/${ bankAccountId }/balances`)
      .then(result => {
        dispatch({
          type: FetchBalances.Success,
          payload: new Balance(result.data)
        });
      })
      .catch(error => {
        dispatch(fetchBalancesFailure);
        throw error;
      })
  };
}
