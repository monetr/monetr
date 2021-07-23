import { BOOTSTRAP_LOGIN, LOGOUT } from "shared/authentication/actions";
import AuthenticationState from "shared/authentication/state";
import * as Sentry from "@sentry/browser";

export default function reducer(state = new AuthenticationState(), action) {
  switch (action.type) {
    case BOOTSTRAP_LOGIN:
      const accountId = action.payload.user.accountId.toString(10)
      Sentry.setUser({
        id: accountId,
        username: `account:${ accountId }`
      });
      return state.merge({
        ...action.payload,
      });
    case LOGOUT:
      Sentry.configureScope(scope => scope.setUser(null));
      return state.merge({
        isAuthenticated: false,
        token: null,
        user: null,
      });
    default:
      return state;
  }
}
