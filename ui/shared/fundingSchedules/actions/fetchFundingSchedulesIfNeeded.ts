import FundingSchedule from "models/FundingSchedule";
import { Map } from 'immutable';
import { getSelectedBankAccountId } from "shared/bankAccounts/selectors/getSelectedBankAccountId";
import { FetchFundingSchedules } from "shared/fundingSchedules/actions";
import request from "shared/util/request";
import { AppDispatch, AppState } from 'store';

interface ActionWithState {
  (dispatch: AppDispatch, getState: () => AppState): Promise<void>
}

export function fetchFundingSchedulesIfNeeded(): ActionWithState {
  return (dispatch, getState) => {
    const selectedBankAccountId = getSelectedBankAccountId(getState());
    if (!selectedBankAccountId) {
      // If the user does not have a bank account selected, then there are no transactions we can request.
      return Promise.resolve();
    }

    dispatch({
      type: FetchFundingSchedules.Request,
    });

    return request()
      .get(`/bank_accounts/${ selectedBankAccountId }/funding_schedules`)
      .then(result => {
        dispatch({
          type: FetchFundingSchedules.Success,
          payload: Map<number, Map<number, FundingSchedule>>().withMutations(map => {
            result.data.map(item => {
              const fundingSchedule = new FundingSchedule(item);
              map.setIn([fundingSchedule.bankAccountId, fundingSchedule.fundingScheduleId], fundingSchedule);
            });
          }),
        });
      })
      .catch(error => {
        dispatch({
          type: FetchFundingSchedules.Failure,
        });

        throw error;
      });
  };
}
