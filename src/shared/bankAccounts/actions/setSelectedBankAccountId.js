import { CHANGE_BANK_ACCOUNT } from "shared/bankAccounts/actions";

export default function setSelectedBankAccountId(bankAccountId = 0) {
  return dispatch => {
    dispatch({
      type: CHANGE_BANK_ACCOUNT,
      payload: bankAccountId,
    });
  };
}
