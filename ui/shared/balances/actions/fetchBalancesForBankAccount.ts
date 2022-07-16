import Balance from 'models/Balance';
import { FetchBalances } from 'shared/balances/actions';
import request from 'shared/util/request';
import { AppActionWithState, AppDispatch, GetAppState } from 'store';

const fetchBalancesRequest = {
  type: FetchBalances.Request,
};

const fetchBalancesFailure = {
  type: FetchBalances.Failure,
};

export default function fetchBalancesForBankAccount(bankAccountId: number | null): AppActionWithState<Promise<void>> {
  return (dispatch: AppDispatch, getState: GetAppState) => {
    if (!bankAccountId) {
      return Promise.resolve();
    }

    dispatch(fetchBalancesRequest);

    return request().get(`/bank_accounts/${ bankAccountId }/balances`)
      .then(result => {
        dispatch({
          type: FetchBalances.Success,
          payload: new Balance(result.data),
        });
      })
      .catch(error => {
        dispatch(fetchBalancesFailure);
        throw error;
      });
  };
}
