import {BOOTSTRAP_LOGIN, LOGIN_SUCCESS, SET_TOKEN} from "./actions";
import AuthenticationState from "./state";


export default function reducer(state = new AuthenticationState(), action) {
  switch (action.type) {
    case BOOTSTRAP_LOGIN:
      return state.merge({
        ...action.payload,
      })
  }
  return state;
}
