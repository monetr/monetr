import { createSelector } from 'reselect';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import { Map } from 'immutable';
import Spending from 'data/Spending';
import { AppState } from 'store';

const getExpensesByBankAccount = (state: AppState): Map<number, Map<number, Spending>> => state.spending.items;

export const getSpending = createSelector<any, any, Map<number, Spending>>(
  [getSelectedBankAccountId, getExpensesByBankAccount],
  (selectedBankAccountId: number, byBankAccount: Map<number, Map<number, Spending>>) => {
    return byBankAccount.get(selectedBankAccountId) ?? Map<number, Spending>();
  },
);
