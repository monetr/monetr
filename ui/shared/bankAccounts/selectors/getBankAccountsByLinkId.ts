import { Map } from 'immutable';
import BankAccount from 'models/BankAccount';
import { createSelector } from 'reselect';
import { getBankAccounts } from 'shared/bankAccounts/selectors/getBankAccounts';

export const getBankAccountsByLinkId = (linkId: number) => createSelector<any, any, Map<number, BankAccount>>(
  getBankAccounts,
  (bankAccounts: Map<number, BankAccount>) => bankAccounts.filter(bankAccount => bankAccount.linkId === linkId)
);
