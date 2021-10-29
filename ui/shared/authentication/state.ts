import User from 'models/User';

export default class AuthenticationState {
  isAuthenticated: boolean;
  isActive: boolean;
  token: string | null;
  user: User | null;
}
