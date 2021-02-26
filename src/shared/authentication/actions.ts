export const LOGIN_REQUEST = 'LOGIN_REQUEST';
export const LOGIN_FAILURE = 'LOGIN_FAILURE';
export const LOGIN_SUCCESS = 'LOGIN_SUCCESS';

export const SET_TOKEN = 'SET_TOKEN';
export const BOOTSTRAP_LOGIN = 'BOOTSTRAP_LOGIN';
export const LOGOUT = 'LOGOUT';

export interface Logout {
  type: typeof LOGOUT,
}
