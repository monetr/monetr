import User from 'models/User';

export default class AuthenticationState {
  isAuthenticated: boolean;
  isActive: boolean;
  user: User | null;
}
