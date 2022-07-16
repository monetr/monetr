import { Map } from 'immutable';
import BankAccount from 'models/BankAccount';
import { AppState } from 'store';

export const getBankAccounts = (state: AppState): Map<number, BankAccount> => state.bankAccounts.items;
