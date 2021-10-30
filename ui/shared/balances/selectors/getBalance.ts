import Balance from 'models/Balance';
import { createSelector } from 'reselect';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import { getBalances } from "shared/balances/selectors/getBalances";

export const getBalance = createSelector<any, any, Balance|null>(
  [getSelectedBankAccountId, getBalances],
  (selectedBankAccountId, balances) => {
    return balances.get(selectedBankAccountId, null);
  },
);

