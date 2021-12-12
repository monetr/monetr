import Balance from 'models/Balance';
import { Logout } from 'shared/authentication/actions';
import { FetchBalances } from 'shared/balances/actions';
import { getBalances } from 'shared/balances/selectors/getBalances';
import BalancesState from 'shared/balances/state';
import { configureStore } from 'store';

describe('balance reducer', () => {

  it('reduce FetchBalances.Success', () => {
    const store = configureStore();

    expect(getBalances(store.getState()).toArray()).toHaveLength(0);

    store.dispatch({
      type: FetchBalances.Success,
      payload: new Balance({
        current: 100,
      })
    });

    expect(getBalances(store.getState()).toArray()).toHaveLength(1);
  });

  it('will handle logout', () => {
    const initialBalanceState = new BalancesState();
    initialBalanceState.items = initialBalanceState.items.set(1, new Balance({
      current: 100,
    }));
    const store = configureStore({
      balances: initialBalanceState,
    });

    expect(getBalances(store.getState()).toArray()).toHaveLength(1);

    store.dispatch({
      type: Logout.Success,
    });

    expect(getBalances(store.getState()).toArray()).toHaveLength(0);
  });

});
