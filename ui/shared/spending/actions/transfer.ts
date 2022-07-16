import Balance from 'models/Balance';
import Spending from 'models/Spending';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import { Transfer } from 'shared/spending/actions';
import request from 'shared/util/request';
import { AppDispatch, AppState } from 'store';

interface ActionWithState {
  (dispatch: AppDispatch, getState: () => AppState): Promise<void>
}

export default function transfer(from: number | null, to: number | null, amount: number): ActionWithState {
  return (dispatch, getState) => {
    if (!from && !to) {
      throw 'must specify a from or a to';
    }

    const selectedBankAccountId = getSelectedBankAccountId(getState());
    if (!selectedBankAccountId) {
      return Promise.resolve();
    }

    return request()
      .post(`/bank_accounts/${ selectedBankAccountId }/spending/transfer`, {
        fromSpendingId: from,
        toSpendingId: to,
        amount: amount,
      })
      .then(result => {
        dispatch({
          type: Transfer,
          payload: {
            balance: new Balance(result.data.balance),
            spending: result.data.spending.map(item => new Spending(item)),
          },
        });
      })
      .catch(error => {
        throw error;
      });
  };
}
