import Spending from 'data/Spending';
import Transaction from 'data/Transaction';
import { Dispatch } from 'redux';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import { UpdateTransaction } from 'shared/transactions/actions';
import request from 'shared/util/request';

interface GetState {
  (): object
}

interface ActionWithState {
  (dispatch: Dispatch, getState: GetState): Promise<void>
}

export default function updateTransaction(transaction: Transaction): ActionWithState {
  return (dispatch, getState) => {
    const selectedBankAccountId = getSelectedBankAccountId(getState());
    if (!selectedBankAccountId) {
      return Promise.resolve();
    }

    if (!transaction.transactionId) {
      return Promise.reject('Transaction must have an Id to be updated');
    }

    transaction.bankAccountId = selectedBankAccountId;

    dispatch({
      type: typeof UpdateTransaction.Request
    });

    return request().put(`/bank_accounts/${ selectedBankAccountId }/transactions/${ transaction.transactionId }`, transaction)
      .then(result => {
        dispatch({
          type: UpdateTransaction.Success,
          payload: {
            transaction: new Transaction(result.data.transaction),
            spending: result.data.spending?.map(item => new Spending(item)),
          }
        });
      })
      .catch(error => {
        dispatch({
          type: typeof UpdateTransaction.Failure
        });

        throw error;
      })
  };
}
