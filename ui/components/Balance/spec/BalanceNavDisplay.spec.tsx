import BalanceNavDisplay from 'components/Balance/BalanceNavDisplay';
import Balance from 'models/Balance';
import React from 'react';
import BalancesState from 'shared/balances/state';
import { CHANGE_BANK_ACCOUNT } from 'shared/bankAccounts/actions';
import { configureStore } from 'store';
import testRenderer from 'testutils/renderer';
import { screen } from '@testing-library/react';

describe('Balance Nav Display', () => {

  it('will not render without balance', () => {
    testRenderer(<BalanceNavDisplay/>);
    expect(screen.queryByTestId('safe-to-spend')).not.toBeInTheDocument();
  });

  it('will render with balances', () => {
    const balances = new BalancesState();
    balances.items = balances.items.set(1, new Balance({
      bankAccountId: 1,
      current: 1100,
      available: 1050,
      safe: 1000,
    }));

    const store = configureStore({
      balances: balances,
    });

    store.dispatch({
      type: CHANGE_BANK_ACCOUNT,
      payload: 1,
    });

    testRenderer(<BalanceNavDisplay/>, {
      store,
    });

    expect(screen.queryByTestId('safe-to-spend')).toBeInTheDocument();
    expect(screen.queryByTestId('safe-to-spend').textContent).toEqual('Safe-To-Spend: $10.00');
  });

});
