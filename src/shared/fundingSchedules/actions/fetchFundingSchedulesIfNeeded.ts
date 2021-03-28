import { Dispatch } from "redux";
import { getSelectedBankAccountId } from "shared/bankAccounts/selectors/getSelectedBankAccountId";
import request from "shared/util/request";

interface GetState {
  (): object
}

interface ActionWithState {
  (dispatch: Dispatch, getState: GetState): Promise<void>
}

export function fetchFundingSchedulesIfNeeded(): ActionWithState {
  return (dispatch, getState) => {
    const selectedBankAccountId = getSelectedBankAccountId(getState());
    if (!selectedBankAccountId) {
      // If the user does not have a bank account selected, then there are no transactions we can request.
      return Promise.resolve();
    }

    return request()
      .get(`/bank_accounts/${ selectedBankAccountId }/funding_schedules`)
      .then(result => {

      })
      .catch(error => {

      });
  };
}
