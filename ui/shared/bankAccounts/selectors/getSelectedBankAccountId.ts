import { AppState } from 'store';

export const getSelectedBankAccountId = (state: AppState): number | null => {
  return state.bankAccounts.selectedBankAccountId || +window.localStorage.getItem('selectedBankAccountId') || state.bankAccounts.items.first()?.bankAccountId;
};
