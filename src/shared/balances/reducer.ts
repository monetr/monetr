import { LOGOUT } from 'shared/authentication/actions';
import { BalanceActions, FetchBalances } from 'shared/balances/actions';
import BalancesState from 'shared/balances/state';


export default function reducer(state: BalancesState = new BalancesState(), action: BalanceActions): BalancesState {
  switch (action.type) {
    case FetchBalances.Success:
      return {
        ...state,
        items: state.items.set(action.payload.bankAccountId, action.payload),
      };
    case LOGOUT:
      return new BalancesState();
    default:
      return state;
  }
}
