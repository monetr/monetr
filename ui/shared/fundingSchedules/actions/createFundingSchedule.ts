import FundingSchedule from "models/FundingSchedule";
import { Dispatch } from "redux";
import { CreateFundingSchedule } from "shared/fundingSchedules/actions";
import request from "shared/util/request";


export default function createFundingSchedule(fundingSchedule: FundingSchedule) {
  return (dispatch: Dispatch) => {
    if (fundingSchedule.bankAccountId <= 0) {
      throw "funding schedule must have a bank account Id";
    }

    dispatch({
      type: CreateFundingSchedule.Request,
    });

    return request()
      .post(`/bank_accounts/${ fundingSchedule.bankAccountId }/funding_schedules`, fundingSchedule)
      .then(result => {
        const newFundingSchedule = new FundingSchedule(result.data);
        dispatch({
          type: CreateFundingSchedule.Success,
          payload: newFundingSchedule,
        });

        return Promise.resolve(newFundingSchedule);
      })
      .catch(error => {
        dispatch({
          type: CreateFundingSchedule.Failure,
        });

        throw error;
      })
  }
}
