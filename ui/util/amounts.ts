/**
 * amountToFriendly takes an amount as it is stored in the API and database and converts it to the amount that is used
 * in the UI. Amounts are stored in their smallest unit. For example; USD is stored in cents. This way the amounts are
 * stored in whole numbers, but are then converted back to dollars and fractions of a dollar in the UI.
 *
 * @param {number} amount The amount as stored by the API, this will be in the smallest possible units for the provided
 *                        currency code. For example. `USD` would be stored in cents instead of dollars and fractions of
 *                        a dollar.
 * @param {string} locale The local code for the current user's perspective. Defaults to `en-US`.
 * @param {string} currency The ISO currency code of the current the amount is in. Defaults to `USD`.
 */
export function amountToFriendly(amount: number, locale: string = 'en-US', currency: string = 'USD'): number {
  const specs = new Intl.NumberFormat(
    locale,
    {
      style: 'currency',
      currency: currency,
    },
  ).resolvedOptions();

  // Determine the multiplier by how many decimal places the final unit would have. For example USD would have 2 decimal
  // places so this would be 10^2 or 100. Where as JPY would have 0 because it is already in the smallest increment.
  // This results in 10^0 which is 1. And the amount/1 remains the same.
  const modifier = Math.pow(10, specs.maximumFractionDigits);

  // Shift the amount over the correct number of decimal places.
  const adjusted = Math.fround(amount / modifier);

  // Truncate any additional decimal places that may exist.
  return +(adjusted.toFixed(specs.maximumFractionDigits));
}

/**
 * friendlyToAmount takes an amount in the regular user friendly form that is displayed and interacted with in the UI
 * and converts it to its smallest unit. For dollars this is converting it into the total number of cents to represent
 * the amount. This is so that the amount is always stored as a whole number and not a decimal.
 *
 * @param {number} friendly The nicely formated (likely decimal) representation of an amount to be converted.
 * @param {string} locale The locale code for the current user's perspective. Defaults to `en-US`.
 * @param {string} currency The ISO currency code of the currency the amount is in. Defaults to `USD`.
 */
export function friendlyToAmount(friendly: number, locale: string = 'en-US', currency: string = 'USD'): number {
  const specs = new Intl.NumberFormat(
    locale,
    {
      style: 'currency',
      currency: currency,
    },
  ).resolvedOptions();

  // Determine the multiplier by how many decimal places the final unit would have. For example USD would have 2 decimal
  // places so this would be 10^2 or 100. Where as JPY would have 0 because it is already in the smallest increment.
  // This results in 10^0 which is 1. And the amount/1 remains the same.
  const modifier = Math.pow(10, specs.maximumFractionDigits);

  // Instead of fractional rounding we want to do whole rounding for storage. Take the friendly amount and multiply it
  // by the modifier based on the number of decimal places the unit has in order to reduce it to it's smallest unit.
  const adjusted = Math.round(friendly * modifier);

  // Truncate any possible decimal places.
  return +(adjusted.toFixed(0));
}

export enum AmountType {
  // Stored amounts are in their smallest unit and may not be user friendly to display in this format.
  Stored,
  // Friendly amounts have been converted into an amount that is displayable and respects the international number
  // format for the currency the amount is in.
  Friendly,
}

/**
 * formatAmount takes the provided number and converts it into a properly formatted string based on the international
 * currency format. If that number is a stored amount it will transform it before formatting. If the number is not a
 * stored amount then the value is not modified at all before formatting.
 *
 * @param {number} amount The amount of money to be formatted, in stored or friendly format.
 * @param {AmountType} type The type of amount value provided. If the value is directly from the API then this would be
 *                          a `AmountType.Stored`. If the value is derived from user input or from somewhere in the UI
 *                          then it is likely `AmountType.Friendly`.
 * @param {string} locale The locale code for the current user's perspective. Defaults to `en-US`.
 * @param {string} currency The ISO currency code of the currency the amount is in. Defaults to `USD`.
 */
export function formatAmount(
  amount: number,
  type: AmountType = AmountType.Stored,
  locale: string = 'en-US',
  currency: string = 'USD',
): string {
  const intl = new Intl.NumberFormat(
    locale,
    {
      style: 'currency',
      currency: currency,
    },
  );

  let value: number = amount;

  // If the provided value is not a friendly one, then convert it into a friendly value.
  if (type === AmountType.Stored) {
    // If we need to conver the value we need to know what the maximumFractionDigits are for the currency.
    const specs = intl.resolvedOptions();

    // Determine the multiplier by how many decimal places the final unit would have. For example USD would have 2
    // decimal places so this would be 10^2 or 100. Where as JPY would have 0 because it is already in the smallest
    // increment. This results in 10^0 which is 1. And the amount/1 remains the same.
    const modifier = Math.pow(10, specs.maximumFractionDigits);

    // Shift the amount over the correct number of decimal places.
    const adjusted = Math.fround(amount / modifier);

    // Truncate any additional decimal places that may exist.
    value = +(adjusted.toFixed(specs.maximumFractionDigits));
  }

  // Convert the resulting friendly amount value into a properly formatted string.
  return intl.format(value);
}
