import Transaction from "data/Transaction";
import { OrderedMap } from "immutable";
import { LOGOUT } from "shared/authentication/actions";
import { CHANGE_BANK_ACCOUNT } from "shared/bankAccounts/actions";
import {
  CHANGE_SELECTED_TRANSACTION,
  FETCH_TRANSACTIONS_FAILURE,
  FETCH_TRANSACTIONS_REQUEST,
  FETCH_TRANSACTIONS_SUCCESS,
  TransactionActions
} from "shared/transactions/actions";
import TransactionState from "shared/transactions/state";

export default function reducer(state: TransactionState = new TransactionState(), action: TransactionActions): TransactionState {
  switch (action.type) {
    case FETCH_TRANSACTIONS_REQUEST:
      return {
        ...state,
        loading: true,
      };
    case FETCH_TRANSACTIONS_FAILURE:
      return {
        ...state,
        loading: false,
      };
    case FETCH_TRANSACTIONS_SUCCESS:
      const mergedTransactions = state.items.set(
        action.bankAccountId,
        state.items.get(action.bankAccountId, OrderedMap<number, Transaction>()).withMutations(map => {
          // We need to do this because ordered maps maintain the order in which objects were set in. By taking the
          // existing ordered map and setting each item we just received, we will either update the existing item or add
          // our new item to the end of the map. Preserving the order we want.
          action.payload.forEach(item => {
            map = map.set(item.transactionId, item);
          })
        })
      );
      return {
        ...state,
        loading: false,
        loaded: true,
        items: mergedTransactions,
      };
    case CHANGE_SELECTED_TRANSACTION:
      return {
        ...state,
        // The comparison logic will allow the selected transaction to be toggled if it is attempted to be selected more
        // than once. Basically if the user clicks a transaction that's already selected then it will unselect it.
        selectedTransactionId: state.selectedTransactionId === action.transactionId ? null : action.transactionId,
      };
    case CHANGE_BANK_ACCOUNT:
      return {
        ...state,
        selectedTransactionId: null,
      };
    case LOGOUT:
      return new TransactionState();
    default:
      return state;
  }
}
