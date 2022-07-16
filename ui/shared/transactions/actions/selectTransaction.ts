import { Dispatch } from 'redux';
import { CHANGE_SELECTED_TRANSACTION } from 'shared/transactions/actions';

export default function selectTransaction(transactionId: number) {
  return (dispatch: Dispatch) => {
    dispatch({
      type: CHANGE_SELECTED_TRANSACTION,
      transactionId: transactionId,
    });
  };
}
