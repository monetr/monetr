import Balance from 'data/Balance';
import { createSelector } from 'reselect';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';

const balancesByBankAccount = state => state.balances.items;

export const getBalance = createSelector<any, any, Balance|null>(
  [getSelectedBankAccountId, balancesByBankAccount],
  (selectedBankAccountId, balances) => {
    return balances.get(selectedBankAccountId, null);
  },
);

