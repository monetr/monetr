import { State } from 'store';

export const getSelectedBankAccountId = (state: State): number | null => {
  return state.bankAccounts.selectedBankAccountId || +window.localStorage.getItem('selectedBankAccountId') || state.bankAccounts.items.first()?.bankAccountId;
};
