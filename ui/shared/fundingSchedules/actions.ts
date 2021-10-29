import FundingSchedule from "models/FundingSchedule";
import { Map } from 'immutable';
import { Logout } from "shared/authentication/actions";
import { RemoveLinkSuccess } from "shared/links/actions";

export enum FetchFundingSchedules {
  Request = 'FetchFundingSchedulesRequest',
  Failure = 'FetchFundingSchedulesFailure',
  Success = 'FetchFundingSchedulesSuccess',
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
  Request = 'CreateFundingScheduleRequest',
  Failure = 'CreateFundingScheduleFailure',
  Success = 'CreateFundingScheduleSuccess',
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
  | RemoveLinkSuccess
  | Logout
