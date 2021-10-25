import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import fetchBalancesForBankAccount from 'shared/balances/actions/fetchBalancesForBankAccount';
import { Dispatch, State } from 'store';

export default function fetchBalances() {
  return (dispatch: Dispatch, getState: () => State) => {
    const selectedBankAccountId = getSelectedBankAccountId(getState());
    if (!selectedBankAccountId) {
      return Promise.resolve();
    }

    return fetchBalancesForBankAccount(selectedBankAccountId)(dispatch, getState);
  };
}
