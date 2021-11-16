import { Logout } from 'shared/authentication/actions';
import { BalanceActions, FetchBalances } from 'shared/balances/actions';
import BalancesState from 'shared/balances/state';
import { Transfer } from 'shared/spending/actions';

export default function reducer(state: BalancesState = new BalancesState(), action: BalanceActions): BalancesState {
  switch (action.type) {
    case FetchBalances.Success:
      return {
        ...state,
        items: state.items.set(action.payload.bankAccountId, action.payload),
      };
    case Transfer:
      return {
        ...state,
        items: state.items.set(action.payload.balance.bankAccountId, action.payload.balance),
      };
    case Logout.Success:
      return new BalancesState();
    default:
      return state;
  }
}
