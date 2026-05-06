package config

type Features struct {
	// TransactionImports is currently in development and so it is tucked behind a
	// basic feature flag here in order to prevent its use while I'm still working
	// on it. Transaction imports are different from transaction uploads. This
	// flag defaults to `false`.
	TransactionImports bool `yaml:"transactionImports"`
}
