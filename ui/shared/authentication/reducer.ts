import { ACTIVATE_SUBSCRIPTION, AuthenticationActions, Login, Logout } from 'shared/authentication/actions';
import AuthenticationState from 'shared/authentication/state';
import * as Sentry from '@sentry/browser';

export default function reducer(state = new AuthenticationState(), action: AuthenticationActions): AuthenticationState {
  switch (action.type) {
    case Login.Pending:
    case Login.Failure:
      return {
        ...state,
        isActive: false,
        isAuthenticated: false,
      };
    case Login.Success:
      // If the user is null then we are not logged in, just return the state.
      if (action.payload) {
        const accountId = action.payload.user.accountId.toString(10);
        Sentry.setUser({
          id: accountId,
          username: `account:${ accountId }`
        });
      }

      return {
        isAuthenticated: true,
        ...action.payload,
      };
    case ACTIVATE_SUBSCRIPTION:
      return {
        ...state,
        isActive: true,
      };
    case Logout.Pending:
    case Logout.Failure:
    case Logout.Success:
      Sentry.setUser(null);
      return new AuthenticationState();
    default:
      return state;
  }
}
