import { Dispatch } from 'redux';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import fetchBalancesForBankAccount from "shared/balances/actions/fetchBalancesForBankAccount";

export default function fetchBalances() {
  return (dispatch: Dispatch, getState) => {
    const selectedBankAccountId = getSelectedBankAccountId(getState());
    if (!selectedBankAccountId) {
      return Promise.resolve();
    }

    return fetchBalancesForBankAccount(selectedBankAccountId)(dispatch, getState);
  };
}
