import User from 'data/User';

export default class AuthenticationState {
  isAuthenticated: boolean;
  isActive: boolean;
  token: string | null;
  user: User | null;
}
