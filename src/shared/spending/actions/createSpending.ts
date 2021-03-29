import Spending from 'data/Spending';
import { Dispatch } from 'redux';
import { CreateSpending } from 'shared/spending/actions';
import request from 'shared/util/request';

export default function createSpending(spending: Spending) {
  return (dispatch: Dispatch) => {
    if (spending.bankAccountId <= 0) {
      throw "spending must have a bank account Id";
    }

    dispatch({
      type: CreateSpending.Request,
    });

    return request()
      .post(`/bank_accounts/${ spending.bankAccountId }/spending`, spending)
      .then(result => {
        dispatch({
          type: CreateSpending.Success,
          payload: new Spending(result.data),
        });
      })
      .catch(error => {
        dispatch({
          type: CreateSpending.Failure,
        });

        throw error;
      });
  };
}
