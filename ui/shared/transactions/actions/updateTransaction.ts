import Balance from 'data/Balance';
import Spending from 'data/Spending';
import Transaction from 'data/Transaction';
import { FetchBalances } from 'shared/balances/actions';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import { UpdateTransaction } from 'shared/transactions/actions';
import request from 'shared/util/request';
import { Dispatch, State } from 'store';

interface ActionWithState {
  (dispatch: Dispatch, getState: () => State): Promise<void>
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
      type: UpdateTransaction.Request
    });

    return request().put(`/bank_accounts/${ selectedBankAccountId }/transactions/${ transaction.transactionId }`, transaction)
      .then(result => {
        // TODO Use multiple redux actions to handle the transaction update and the spending update.
        dispatch({
          type: UpdateTransaction.Success,
          payload: {
            transaction: new Transaction(result.data.transaction),
            spending: result.data.spending?.map(item => new Spending(item)),
          }
        });

        if (result.data.balance) {
          dispatch({
            type: FetchBalances.Success,
            payload: new Balance(result.data.balance),
          });
        }
      })
      .catch(error => {
        dispatch({
          type: UpdateTransaction.Failure
        });

        throw error;
      })
  };
}
