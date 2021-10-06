import { createSelector } from "reselect";
import { getBankAccounts } from "shared/bankAccounts/selectors/getBankAccounts";
import { Map } from 'immutable';
import BankAccount from "data/BankAccount";

export const getBankAccountsByLinkId = (linkId: number) => createSelector<any, any, Map<number, BankAccount>>(
  [getBankAccounts],
  bankAccounts => bankAccounts.filter(bankAccount => bankAccount.linkId === linkId)
);
