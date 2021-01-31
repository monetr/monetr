import {applyMiddleware, createStore, compose, Action} from 'redux';
import {composeWithDevTools} from 'redux-devtools-extension';
import reducers from './shared/state';
import Cookies from 'js-cookie';
import thunk, {ThunkAction, ThunkDispatch} from 'redux-thunk';
import AuthenticationState from "./shared/authentication/state";
import {Map} from "immutable";
import BootstrapState from "./shared/bootstrap/state";


export default function configureStore(initialState = {
  authentication: new AuthenticationState(),
  bootstrap: new BootstrapState(),
}) {
  const composeEnhancer = process.env.NODE_ENV !== 'production' ? composeWithDevTools({
    name: 'Primary',
    maxAge: 150,
    trace: true,
    traceLimit: 25,
  }) : compose;

  let store = createStore(
    reducers,
    initialState,
    composeEnhancer(applyMiddleware(thunk)),
  );

  return store;
}
