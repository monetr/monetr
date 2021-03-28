import { LOGOUT } from "shared/authentication/actions";
import { CreateFundingSchedule, FundingScheduleActions } from "shared/fundingSchedules/actions";
import FundingScheduleState from "shared/fundingSchedules/state";


export default function reducer(state: FundingScheduleState = new FundingScheduleState(), action: FundingScheduleActions): FundingScheduleState {
  switch (action.type) {
    case CreateFundingSchedule.Success:
      return {
        ...state
      };
    case CreateFundingSchedule.Failure:
    case LOGOUT:
      return new FundingScheduleState();
    default:
      return state;
  }
}
