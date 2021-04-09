import Balance from 'data/Balance';
import { Logout } from 'shared/authentication/actions';


export enum FetchBalances {
  Request = 'FetchBalancesRequest',
  Failure = 'FetchBalancesFailure',
  Success = 'FetchBalancesSuccess',
}

export interface FetchBalancesRequest {
  type: typeof FetchBalances.Request;
}

export interface FetchBalancesFailure {
  type: typeof FetchBalances.Failure;
}

export interface FetchBalancesSuccess {
  type: typeof FetchBalances.Success;
  payload: Balance;
}

export type BalanceActions =
  FetchBalancesRequest
  | FetchBalancesFailure
  | FetchBalancesSuccess
  | Logout
