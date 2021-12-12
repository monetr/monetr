import Balance from 'models/Balance';
import { FetchBalances } from 'shared/balances/actions';
import { getBalance } from 'shared/balances/selectors/getBalance';
import { getBalances } from 'shared/balances/selectors/getBalances';
import { configureStore } from 'store';

describe('getBalances', () => {

  it('will return the balances in the store', () => {
    const store = configureStore();

    { // Make sure that the initial state is empty.
      const balances = getBalances(store.getState());
      expect(balances.isEmpty()).toBeTruthy();
    }

    { // Now seed the store.
      store.dispatch({
        type: FetchBalances.Success,
        payload: new Balance({
          bankAccountId: 1,
        })
      });
    }

    { // Now make sure we can retrieve the balance.
      const balances = getBalances(store.getState());
      expect(balances.has(1)).toBeTruthy();
    }

  });

});
