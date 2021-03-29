import { LOGOUT } from "shared/authentication/actions";
import { CreateFundingSchedule, FetchFundingSchedules, FundingScheduleActions } from "shared/fundingSchedules/actions";
import FundingScheduleState from "shared/fundingSchedules/state";

export default function reducer(state: FundingScheduleState = new FundingScheduleState(), action: FundingScheduleActions): FundingScheduleState {
  switch (action.type) {
    case CreateFundingSchedule.Request:
    case FetchFundingSchedules.Request:
      return {
        ...state,
        loading: true,
      };
    case FetchFundingSchedules.Failure:
    case CreateFundingSchedule.Failure:
      return {
        ...state,
        loading: false,
      };
    case FetchFundingSchedules.Success:
      return {
        ...state,
        items: state.items.mergeDeep(action.payload),
      };
    case CreateFundingSchedule.Success:
      // With create funding schedule the payload will be a single funding schedule that was just created. So we can
      // just do a set in for the current items to add the new schedule.
      return {
        ...state,
        loading: false,
        items: state.items.setIn([
          action.payload.bankAccountId,
          action.payload.fundingScheduleId,
        ], action.payload),
      };
    case LOGOUT:
      return new FundingScheduleState();
    default:
      return state;
  }
}
