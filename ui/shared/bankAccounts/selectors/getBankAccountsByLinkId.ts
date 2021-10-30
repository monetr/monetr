import { createSelector } from 'reselect';
import { getBankAccounts } from 'shared/bankAccounts/selectors/getBankAccounts';
import { Map } from 'immutable';
import BankAccount from 'models/BankAccount';

export const getBankAccountsByLinkId = (linkId: number) => createSelector<any, any, Map<number, BankAccount>>(
  getBankAccounts,
  (bankAccounts: Map<number, BankAccount>) => bankAccounts.filter(bankAccount => bankAccount.linkId === linkId)
);
