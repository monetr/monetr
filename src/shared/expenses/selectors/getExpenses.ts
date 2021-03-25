import { createSelector } from "reselect";
import { getSelectedBankAccountId } from "shared/bankAccounts/selectors/getSelectedBankAccountId";
import { Map } from "immutable";
import Expense from "data/Expense";

const getExpensesByBankAccount = state => state.expenses.items;

export const getExpenses = createSelector<any, any, Map<number, Expense>>(
  [getSelectedBankAccountId, getExpensesByBankAccount],
  (selectedBankAccountId: number, byBankAccount: Map<number, Map<number, Expense>>) => {
    return byBankAccount.get(selectedBankAccountId) ?? Map<number, Expense>();
  },
);
