import { applyMiddleware, combineReducers, compose, createStore } from 'redux';
import * as Sentry from '@sentry/react';

import { composeWithDevTools } from 'redux-devtools-extension';
import thunk from 'redux-thunk';
import authentication from 'shared/authentication/reducer';
import balances from 'shared/balances/reducer';
import bankAccounts from 'shared/bankAccounts/reducer';
import fundingSchedules from 'shared/fundingSchedules/reducer';
import links from 'shared/links/reducer';
import spending from 'shared/spending/reducer';
import transactions from 'shared/transactions/reducer';

export const reducers = combineReducers({
  authentication,
  balances,
  bankAccounts,
  fundingSchedules,
  links,
  spending,
  transactions,
});

export type AppState = ReturnType<typeof reducers>;

export function configureStore(initialState?: Partial<AppState>) {
  const composeEnhancer = process.env.NODE_ENV !== 'production' ? composeWithDevTools({
    name: 'Primary',
    maxAge: 150,
    trace: true,
    traceLimit: 25,
  }) : compose;

  const sentryReduxEnhancer = Sentry.createReduxEnhancer({});

  return createStore(
    reducers,
    initialState || {},
    composeEnhancer(compose(applyMiddleware(thunk), sentryReduxEnhancer)),
  );
}

export const store = configureStore();

export type AppDispatch = typeof store.dispatch

export type AppStore = typeof store;

export type GetAppState = () => AppState;

export interface AppAction<T> {
  (dispatch: AppDispatch): T
}

export interface AppActionWithState<T> {
  (dispatch: AppDispatch, getState: GetAppState): T
}

export default store;
