import { BOOTSTRAP_LOGIN, LOGOUT } from "shared/authentication/actions";
import AuthenticationState from "shared/authentication/state";

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
