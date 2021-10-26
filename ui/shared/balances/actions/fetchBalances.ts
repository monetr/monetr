import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import fetchBalancesForBankAccount from 'shared/balances/actions/fetchBalancesForBankAccount';
import { AppDispatch, AppState } from 'store';

export default function fetchBalances() {
  return (dispatch: AppDispatch, getState: () => AppState) => {
    const selectedBankAccountId = getSelectedBankAccountId(getState());
    if (!selectedBankAccountId) {
      return Promise.resolve();
    }

    return fetchBalancesForBankAccount(selectedBankAccountId)(dispatch, getState);
  };
}
