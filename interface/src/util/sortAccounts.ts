import type BankAccount from '@monetr/interface/models/BankAccount';
import type { ID } from '@monetr/interface/models/ID';

/**
 * sortAccounts will take an array of accounts and sort them by the account type and sub type priorities. If an order is
 * provided (the link's bankAccountOrder) then that wins, and anything not in it lands after in the type order.
 */
export default function sortAccounts(
  bankAccounts: Array<BankAccount> | null | undefined,
  order?: Array<ID<BankAccount>> | null,
): Array<BankAccount> {
  if (!bankAccounts) {
    return [];
  }

  // Depository accounts should be the highest value. Account types that are not listed here will
  // have a value of -1.
  const accountTypeOrder = ['loan', 'credit', 'depository'];
  // Checking sub account types should have the highest value. Sub account types that are not
  // listed here will have a value of -1.
  const accountSubTypeOrder = ['money market', 'mortgage', 'auto', 'credit card', 'savings', 'checking'];

  // score returns the sortable weight for a single account. We pulled this out into its own function so we don't have
  // to index into parallel arrays, which noUncheckedIndexedAccess does not like.
  function score(account: BankAccount): number {
    // Put inactive items last.
    const multiplier = account.status === 'inactive' ? -10 : 1;
    let value = accountTypeOrder.indexOf(account.accountType) + 1;
    value += accountSubTypeOrder.indexOf(account.accountSubType) + 1;
    return value * multiplier;
  }

  // I want to sort these in descenging order. So invert whether or not the value returned
  // is negative or positive.
  const sorted = bankAccounts.sort((a, b) => {
    const aValue = score(a);
    const bValue = score(b);
    return aValue < bValue ? 1 : aValue > bValue ? -1 : 0;
  });
  if (!order || order.length === 0) {
    return sorted;
  }

  // Ids in the order that don't match an account (like one that was removed) are just skipped.
  const ordered = order
    .map(bankAccountId => sorted.find(item => item.bankAccountId === bankAccountId))
    .filter(item => item !== undefined);
  return ordered.concat(sorted.filter(item => !order.includes(item.bankAccountId)));
}
