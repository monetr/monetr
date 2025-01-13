package consts

// DefaultCurrencyCode is monetr's default currency code that will be used if an
// actual currency code cannot be derived from the datasource the transactions
// or balances are coming from.
const DefaultCurrencyCode = "USD"

// DefaultLocale is similar to the default currency code, when a locale cannot
// be determined then we fall back to this. This may also be used if a valid
// locale is provided to monetr but monetr does not posses the data to use that
// locale code.
const DefaultLocale = "en_US"
