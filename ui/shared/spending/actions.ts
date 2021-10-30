import Balance from 'models/Balance';
import Spending from "models/Spending";
import { Map } from 'immutable';
import { Logout } from "shared/authentication/actions";
import { ChangeBankAccount } from "shared/bankAccounts/actions";
import { UpdateTransactionSuccess } from 'shared/transactions/actions';

export const FETCH_SPENDING_REQUEST = 'FETCH_SPENDING_REQUEST';
export const FETCH_SPENDING_FAILURE = 'FETCH_SPENDING_FAILURE';
export const FETCH_SPENDING_SUCCESS = 'FETCH_SPENDING_SUCCESS';

export interface FetchSpendingSuccess {
  type: typeof FETCH_SPENDING_SUCCESS;
  payload: Map<number, Map<number, Spending>>;
}

export interface FetchSpendingFailure {
  type: typeof FETCH_SPENDING_FAILURE;
}

export interface FetchSpendingRequest {
  type: typeof FETCH_SPENDING_REQUEST;
}

export enum CreateSpending {
  Request = 'CreateSpendingRequest',
  Failure = 'CreateSpendingFailure',
  Success = 'CreateSpendingSuccess',
}

export interface CreateSpendingRequest {
  type: typeof CreateSpending.Request;
}

export interface CreateSpendingFailure {
  type: typeof CreateSpending.Failure;
}

export interface CreateSpendingSuccess {
  type: typeof CreateSpending.Success;
  payload: Spending;
}

export enum DeleteSpending {
  Request = 'DeleteSpendingRequest',
  Failure = 'DeleteSpendingFailure',
  Success = 'DeleteSpendingSuccess',
}

export interface DeleteSpendingRequest {
  type: typeof DeleteSpending.Request;
}

export interface DeleteSpendingFailure {
  type: typeof DeleteSpending.Failure;
}

export interface DeleteSpendingSuccess {
  type: typeof DeleteSpending.Success;
  payload: Spending;
}

export const SelectExpense = 'SelectExpense';

export interface SelectExpense {
  type: typeof SelectExpense;
  expenseId: number | null;
}

export const SelectGoal = 'SelectGoal';

export interface SelectGoal {
  type: typeof SelectGoal;
  goalId: number | null;
}

export const Transfer = 'Transfer';

export interface Transfer {
  type: typeof Transfer;
  payload: {
    balance: Balance;
    spending: Spending[];
  };
}

export enum UpdateSpending {
  Request = 'UpdateSpendingRequest',
  Failure = 'UpdateSpendingFailure',
  Success = 'UpdateSpendingSuccess',
}

export interface UpdateSpendingSuccess {
  type: typeof UpdateSpending.Success;
  payload: Spending;
}

export type SpendingActions =
  FetchSpendingRequest
  | FetchSpendingFailure
  | FetchSpendingSuccess
  | CreateSpendingRequest
  | CreateSpendingFailure
  | CreateSpendingSuccess
  | DeleteSpendingRequest
  | DeleteSpendingFailure
  | DeleteSpendingSuccess
  | SelectExpense
  | SelectGoal
  | Transfer
  | UpdateTransactionSuccess
  | UpdateSpendingSuccess
  | Logout
  | ChangeBankAccount
