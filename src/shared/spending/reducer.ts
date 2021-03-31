import { LOGOUT } from "shared/authentication/actions";
import { CHANGE_BANK_ACCOUNT } from 'shared/bankAccounts/actions';
import {
  CreateSpending,
  FETCH_SPENDING_FAILURE,
  FETCH_SPENDING_REQUEST,
  FETCH_SPENDING_SUCCESS,
  SelectExpense,
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
    case CHANGE_BANK_ACCOUNT:
      return {
        ...state,
        selectedExpenseId: null,
      };
    case SelectExpense:
      return {
        ...state,
        // The comparison logic will allow the selected expense to be toggled if it is attempted to be selected more
        // than once. Basically if the user clicks a expense that's already selected then it will unselect it.
        selectedExpenseId: state.selectedExpenseId === action.expenseId ? null : action.expenseId,
      };
    case LOGOUT:
      return new SpendingState();
    default:
      return state;
  }
}
