import { Dispatch } from 'redux';
import { SelectExpense } from 'shared/spending/actions';

export default function selectExpense(expenseId: number) {
  return (dispatch: Dispatch) => {
    dispatch({
      type: SelectExpense,
      expenseId: expenseId,
    });
  };
}
