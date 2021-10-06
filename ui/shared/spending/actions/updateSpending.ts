import Spending from "data/Spending";
import { Dispatch } from "redux";
import { UpdateSpending } from "shared/spending/actions";
import request from "shared/util/request";


export default function updateSpending(spending: Spending) {
  return (dispatch: Dispatch) => {
    if (spending.bankAccountId <= 0) {
      throw "spending must have a bank account Id";
    }

    dispatch({
      type: UpdateSpending.Request,
    });

    return request()
      .put(`/bank_accounts/${ spending.bankAccountId }/spending/${ spending.spendingId }`, spending)
      .then(result => {
        dispatch({
          type: UpdateSpending.Success,
          payload: new Spending(result.data),
        });
      })
      .catch(error => {
        dispatch({
          type: UpdateSpending.Failure,
        });

        throw error;
      });
  };
}
