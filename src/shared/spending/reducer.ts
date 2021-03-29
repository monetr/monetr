import { LOGOUT } from "shared/authentication/actions";
import {
  CreateSpending,
  FETCH_SPENDING_FAILURE,
  FETCH_SPENDING_REQUEST,
  FETCH_SPENDING_SUCCESS,
  SpendingActions
} from "shared/spending/actions";
import SpendingState from "shared/spending/state";

export default function reducer(state: SpendingState = new SpendingState(), action: SpendingActions): SpendingState {
  switch (action.type) {
    case CreateSpending.Request:
    case FETCH_SPENDING_REQUEST:
      return {
        ...state,
        loading: true,
      };
    case CreateSpending.Failure:
    case FETCH_SPENDING_FAILURE:
      return {
        ...state,
        loading: false
      };
    case FETCH_SPENDING_SUCCESS:
      return {
        ...state,
        loaded: true,
        loading: false,
        items: state.items.mergeDeep(action.payload),
      };
    case CreateSpending.Success:
      return {
        ...state,
        loading: false,
        items: state.items.setIn([action.payload.bankAccountId, action.payload.spendingId], action.payload),
      };
    case LOGOUT:
      return new SpendingState();
    default:
      return state;
  }
}
