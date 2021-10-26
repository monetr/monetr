import BankAccount from 'data/BankAccount';
import { Map } from 'immutable';
import { AppState } from 'store';

export const getBankAccounts = (state: AppState): Map<number, BankAccount> => state.bankAccounts.items;
