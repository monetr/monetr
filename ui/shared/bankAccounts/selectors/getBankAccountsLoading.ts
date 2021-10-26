import { AppState } from 'store';

export const getBankAccountsLoading = (state: AppState): boolean => state.bankAccounts.loading;
