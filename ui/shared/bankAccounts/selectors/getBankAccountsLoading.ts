import { State } from 'store';

export const getBankAccountsLoading = (state: State): boolean => state.bankAccounts.loading;
