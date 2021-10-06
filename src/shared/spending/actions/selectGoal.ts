import { Dispatch } from 'redux';
import { SelectGoal } from 'shared/spending/actions';

export default function selectGoal(goalId: number) {
  return (dispatch: Dispatch) => {
    dispatch({
      type: SelectGoal,
      goalId: goalId,
    });
  };
}
