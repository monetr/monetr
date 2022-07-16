import BankAccount from 'models/BankAccount';
import fetchBalancesForBankAccount from 'shared/balances/actions/fetchBalancesForBankAccount';
import { getBalances } from 'shared/balances/selectors/getBalances';
import { getBankAccounts } from 'shared/bankAccounts/selectors/getBankAccounts';
import { AppActionWithState, AppDispatch, GetAppState } from 'store';

export default function fetchMissingBankAccountBalances(): AppActionWithState<Promise<void[]>> {
  return (dispatch: AppDispatch, getState: GetAppState): Promise<void[]> => {
    const bankAccounts = getBankAccounts(getState());
    const balances = getBalances(getState());

    const missingPromises: Promise<void>[] = [];
    bankAccounts.forEach((item: BankAccount) => {
      if (balances.has(item.bankAccountId)) {
        return;
      }

      missingPromises.push(fetchBalancesForBankAccount(item.bankAccountId)(dispatch, getState));
    });

    return Promise.all(missingPromises);
  };
}
