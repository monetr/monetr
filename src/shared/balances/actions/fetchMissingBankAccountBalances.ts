import { getSelectedBankAccountId } from "shared/bankAccounts/selectors/getSelectedBankAccountId";
import fetchBalancesForBankAccount from "shared/balances/actions/fetchBalancesForBankAccount";
import { getBankAccounts } from "shared/bankAccounts/selectors/getBankAccounts";
import { Dispatch } from "redux";
import { getBalances } from "shared/balances/selectors/getBalances";
import BankAccount from "data/BankAccount";


export default function fetchMissingBankAccountBalances() {
  return (dispatch: Dispatch, getState) => {
    const bankAccounts = getBankAccounts(getState());
    const balances = getBalances(getState());

    const missingPromises: Promise<void>[] = [];
    bankAccounts.forEach((item: BankAccount) => {
      if (balances.has(item.bankAccountId)) {
        return;
      }

      missingPromises.push(fetchBalancesForBankAccount(item.bankAccountId)(dispatch, getState));
    })

    return Promise.all(missingPromises);
  }
}
