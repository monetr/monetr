import Spending from "models/Spending";
import { Dispatch } from "redux";
import { DeleteSpending } from "shared/spending/actions";
import request from "shared/util/request";


export default function deleteSpending(spending: Spending) {
  return (dispatch: Dispatch) => {
    if (spending.bankAccountId <= 0) {
      throw "spending must have a bank account Id";
    }

    dispatch({
      type: DeleteSpending.Request,
    });

    return request()
      .delete(`/bank_accounts/${ spending.bankAccountId }/spending/${ spending.spendingId }`)
      .then(() => {
        dispatch({
          type: DeleteSpending.Success,
          payload: spending,
        });
      })
      .catch(error => {
        dispatch({
          type: DeleteSpending.Failure,
        });

        throw error;
      });
  };
}
