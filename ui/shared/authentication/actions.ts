import User from 'models/User';
import { Action } from 'redux';

export enum Login {
  Pending = 'LoginPending',
  Failure = 'LoginFailure',
  Success = 'LoginSuccess',
}

export interface LoginPending extends Action<typeof Login.Pending> {
  type: typeof Login.Pending;
}

export interface LoginFailure extends Action<typeof Login.Failure> {
  type: typeof Login.Failure;
}

export interface LoginSuccess extends Action<typeof Login.Success> {
  type: typeof Login.Success;
  payload: {
    user: User;
    isActive: boolean;
    hasSubscription: boolean;
  };
}

export enum Logout {
  Pending = 'LogoutPending',
  Failure = 'LogoutFailure',
  Success = 'LogoutSuccess',
}

export interface LogoutPending extends Action<typeof Logout.Pending> {
  type: typeof Logout.Pending;
}

export interface LogoutFailure extends Action<typeof Logout.Failure> {
  type: typeof Logout.Failure;
}

export interface LogoutSuccess extends Action<typeof Logout.Success> {
  type: typeof Logout.Success;
}

export const ACTIVATE_SUBSCRIPTION = 'ACTIVATE_SUBSCRIPTION';

export interface ActivateSubscription {
  type: typeof ACTIVATE_SUBSCRIPTION;
}

export type LogoutActions = LogoutPending | LogoutFailure | LogoutSuccess;

export type AuthenticationActions =
  LoginPending
  | LoginFailure
  | LoginSuccess
  | LogoutPending
  | LogoutFailure
  | LogoutSuccess
  | ActivateSubscription;
