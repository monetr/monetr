import fetchBalancesForBankAccount from 'shared/balances/actions/fetchBalancesForBankAccount';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import { AppActionWithState, AppDispatch, AppState, GetAppState } from 'store';

export default function fetchBalances(): AppActionWithState<Promise<void>> {
  return (dispatch: AppDispatch, getState: GetAppState): Promise<void> => {
    const selectedBankAccountId = getSelectedBankAccountId(getState());
    if (!selectedBankAccountId) {
      return Promise.resolve();
    }

    return fetchBalancesForBankAccount(selectedBankAccountId)(dispatch, getState);
  };
}
