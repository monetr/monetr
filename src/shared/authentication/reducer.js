import {BOOTSTRAP_LOGIN, LOGOUT} from "./actions";
import AuthenticationState from "./state";

export default function reducer(state = new AuthenticationState(), action) {
  switch (action.type) {
    case BOOTSTRAP_LOGIN:
      return state.merge({
        ...action.payload,
      });
    case LOGOUT:
      return state.merge({
        isAuthenticated: false,
        token: null,
        user: null,
      });
    default:
      return state;
  }
}
