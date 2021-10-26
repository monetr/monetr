import { CHANGE_BANK_ACCOUNT } from 'shared/bankAccounts/actions';
import { AppDispatch } from 'store';

export default function setSelectedBankAccountId(bankAccountId = 0): (dispatch: AppDispatch) => void {
  return (dispatch: AppDispatch) => {
    // Store the selected bank accountId inside local storage. This way we can bring the user right back to their
    // selected bank the next time they load our app.
    window.localStorage.setItem('selectedBankAccountId', bankAccountId.toString(10));

    dispatch({
      type: CHANGE_BANK_ACCOUNT,
      bankAccountId: bankAccountId,
    });
  };
}
