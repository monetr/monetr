import type BankAccount from '@monetr/interface/models/BankAccount';

/**
 * sortAccounts will take an array of accounts and sort them by the account type and sub type priorities.
 */
export default function sortAccounts(bankAccounts: Array<BankAccount> | null | undefined): Array<BankAccount> {
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
  return bankAccounts.sort((a, b) => {
    const aValue = score(a);
    const bValue = score(b);
    return aValue < bValue ? 1 : aValue > bValue ? -1 : 0;
  });
}
