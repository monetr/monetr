import { createSelector } from "reselect";
import { getBankAccounts } from "shared/bankAccounts/selectors/getBankAccounts";

export const getBankAccountsByLinkId = linkId => createSelector(
  [getBankAccounts],
  bankAccounts => bankAccounts.filter(bankAccount => bankAccount.linkId === linkId)
);
