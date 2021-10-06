// @ts-ignore
export const getSelectedBankAccountId = (state: any): number | null => {
  return state?.bankAccounts?.selectedBankAccountId || +window.localStorage.getItem('selectedBankAccountId') || state.bankAccounts.items.first()?.bankAccountId;
};
