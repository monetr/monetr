import { createSelector } from 'reselect';

const getSelectedExpenseId = state => state.spending.selectedExpenseId;

export const getExpenseIsSelected = (expenseId: number) => createSelector<any, any, boolean>(
  [getSelectedExpenseId],
  selectedExpenseId => expenseId === selectedExpenseId,
);
