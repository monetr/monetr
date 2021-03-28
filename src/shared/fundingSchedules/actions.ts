import FundingSchedule from "data/FundingSchedule";
import { Map } from 'immutable';
import { Logout } from "shared/authentication/actions";

export enum FetchFundingSchedules {
  Request,
  Failure,
  Success,
}

export interface FetchFundingSchedulesRequest {
  type: typeof FetchFundingSchedules.Request;
}

export interface FetchFundingSchedulesFailure {
  type: typeof FetchFundingSchedules.Failure;
}

export interface FetchFundingSchedulesSuccess {
  type: typeof FetchFundingSchedules.Success;
  payload: Map<number, Map<number, FundingSchedule>>;
}

export enum CreateFundingSchedule {
  Request,
  Failure,
  Success,
}

export interface CreateFundingScheduleRequest {
  type: typeof CreateFundingSchedule.Request;
}

export interface CreateFundingScheduleFailure {
  type: typeof CreateFundingSchedule.Failure;
}

export interface CreateFundingScheduleSuccess {
  type: typeof CreateFundingSchedule.Success;
  payload: FundingSchedule;
}

export type FundingScheduleActions =
  FetchFundingSchedulesRequest
  | FetchFundingSchedulesFailure
  | FetchFundingSchedulesSuccess
  | CreateFundingScheduleRequest
  | CreateFundingScheduleFailure
  | CreateFundingScheduleSuccess
  | Logout
