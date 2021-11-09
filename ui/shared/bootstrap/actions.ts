import { Action } from 'redux';
import BootstrapState from 'shared/bootstrap/state';

export enum Bootstrap {
  Begin = 'BootstrapBegin',
  Failure = 'BootstrapFailure',
  Success = 'BootstrapSuccess',
}

export interface BootstrapBegin extends Action<typeof Bootstrap.Begin> {
  type: typeof Bootstrap.Begin;
}

export interface BootstrapFailure extends Action<typeof Bootstrap.Failure> {
  type: typeof Bootstrap.Failure;
}

export interface BootstrapSuccess extends Action<typeof Bootstrap.Success> {
  type: typeof Bootstrap.Success;
  payload: Partial<BootstrapState>;
}

export type BootstrapActions = BootstrapBegin | BootstrapFailure | BootstrapSuccess;
