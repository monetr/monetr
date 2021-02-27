import TransactionState from "shared/transactions/state";
import {
  FETCH_TRANSACTIONS_FAILURE,
  FETCH_TRANSACTIONS_REQUEST,
  FETCH_TRANSACTIONS_SUCCESS,
  TransactionActions
} from "shared/transactions/actions";
import { LOGOUT } from "shared/authentication/actions";
import { OrderedMap } from "immutable";
import Transaction from "data/Transaction";

export default function reducer(state: TransactionState = new TransactionState(), action: TransactionActions) {
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
    case LOGOUT:
      return new TransactionState();
    default:
      return state;
  }
}
