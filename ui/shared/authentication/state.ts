import User from 'models/User';

export default class AuthenticationState {
  isAuthenticated: boolean;
  // isActive indicates whether or not the user's subscription is activated. This is determined by the authentication
  // API call, or by the /users/me request.
  isActive: boolean;
  user: User | null;
}
