import { createSelector } from 'reselect';

/**
  * getSelectedExpenseId will return the current expense that the user has selected in the UI or null.
  */
const getSelectedExpenseId = state => state.spending.selectedExpenseId;

/**
 * getExpenseIsSelected will return true if the provided spendingId
 * matches the ID of the currently selected spending object under the
 * "expenses" view. While spendingId is globally unique between
 * expenses and goals.
 *
 * This will only return true if the provided spendingId is an
 * expense and is selected.
 *
 * @returns Whether or not the provided spending object is selected.
 */
export const getExpenseIsSelected = (spendingId: number) => createSelector<any, any, boolean>(
  [getSelectedExpenseId],
  selectedExpenseId => spendingId === selectedExpenseId,
);
