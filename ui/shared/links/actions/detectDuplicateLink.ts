import { Map } from 'immutable';
import BankAccount from 'models/BankAccount';
import Link from 'models/Link';
import { PlaidLinkOnSuccessMetadata } from 'react-plaid-link/src/types/index';
import { AppActionWithState, AppDispatch, AppState, GetAppState } from 'store';

export default function detectDuplicateLink(metadata: PlaidLinkOnSuccessMetadata): AppActionWithState<boolean> {
  return (_: AppDispatch, getState: GetAppState): boolean => {
    const state = getState();

    // Gather all the links that are for the institution that we got from the success metadata.
    const linksForInstitution: Map<number, Link> = state.links.items.filter((item: Link) => {
      return item.plaidInstitutionId === metadata.institution.institution_id;
    });

    // Now that we have all the links that _might_ conflict. Check all the bank accounts for those links. If there is
    // a conflict then there will be at least one item in this array.
    const bankAccounts = state.bankAccounts.items.filter((item: BankAccount) => {
      return linksForInstitution.has(item.linkId) && !!metadata.accounts.find(account => account.mask === item.mask);
    });

    // If there are any; then there is a good chance its a duplicate.
    return !bankAccounts.isEmpty();
  };
}
