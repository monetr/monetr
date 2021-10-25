import BankAccount from 'data/BankAccount';
import { Map } from 'immutable';
import { State } from 'store';

export const getBankAccounts = (state: State): Map<number, BankAccount> => state.bankAccounts.items;
