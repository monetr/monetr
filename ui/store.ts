import { applyMiddleware, combineReducers, compose, createStore } from 'redux';
import { composeWithDevTools } from 'redux-devtools-extension';
import thunk from 'redux-thunk';
import authentication from 'shared/authentication/reducer';
import balances from 'shared/balances/reducer';
import bankAccounts from 'shared/bankAccounts/reducer';
import bootstrap from 'shared/bootstrap/reducer';
import fundingSchedules from 'shared/fundingSchedules/reducer';
import links from 'shared/links/reducer';
import spending from 'shared/spending/reducer';
import transactions from 'shared/transactions/reducer';

export const reducers = combineReducers({
  authentication,
  balances,
  bankAccounts,
  bootstrap,
  fundingSchedules,
  links,
  spending,
  transactions,
});

export type AppState = ReturnType<typeof reducers>;

export function configureStore() {
  const composeEnhancer = process.env.NODE_ENV !== 'production' ? composeWithDevTools({
    name: 'Primary',
    maxAge: 150,
    trace: true,
    traceLimit: 25,
  }) : compose;

  return createStore(
    reducers,
    {},
    composeEnhancer(applyMiddleware(thunk)),
  );
}

export const store = configureStore();

export type AppDispatch = typeof store.dispatch


export default store;