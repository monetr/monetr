import BankAccount from '@monetr/interface/models/BankAccount';

/**
 * sortAccounts will take an array of accounts and sort them by the account type and sub type priorities.
 */
export default function sortAccounts(bankAccounts: Array<BankAccount> | undefined): Array<BankAccount> {
  if (!bankAccounts) {
    return [];
  }

  // Depository accounts should be the highest value. Account types that are not listed here will
  // have a value of -1.
  const accountTypeOrder = ['loan', 'credit', 'depository'];
  // Checking sub account types should have the highest value. Sub account types that are not
  // listed here will have a value of -1.
  const accountSubTypeOrder = ['money market', 'mortgage', 'auto', 'credit card', 'savings', 'checking'];

  const result = bankAccounts.sort((a, b) => {
    const items = [a, b];
    const values = [
      0, // a
      0, // b
    ];
    for (let i = 0; i < 2; i++) {
      // Put inactive items last.
      const multiplier = items[i].status === 'inactive' ? -10 : 1;
      values[i] += accountTypeOrder.indexOf(items[i].accountType) + 1;
      values[i] += accountSubTypeOrder.indexOf(items[i].accountSubType) + 1;
      values[i] *= multiplier;
    }

    // I want to sort these in descenging order. So invert whether or not the value returned
    // is negative or positive.
    return values[0] < values[1] ? 1 : values[0] > values[1] ? -1 : 0;
  });
  return result;
}
