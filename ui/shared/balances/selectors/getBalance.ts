import Balance from 'models/Balance';
import { createSelector } from 'reselect';
import { getBalances } from 'shared/balances/selectors/getBalances';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';

export const getBalance = createSelector<any, any, Balance|null>(
  [getSelectedBankAccountId, getBalances],
  (selectedBankAccountId, balances) => {
    return balances.get(selectedBankAccountId, null);
  },
);

