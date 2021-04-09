import Balance from 'data/Balance';
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

export default function fetchBalances() {
  return (dispatch: Dispatch, getState) => {
    const selectedBankAccountId = getSelectedBankAccountId(getState());
    if (!selectedBankAccountId) {
      return Promise.resolve();
    }

    dispatch(fetchBalancesRequest);

    return request().get(`/bank_accounts/${ selectedBankAccountId }/balances`)
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
