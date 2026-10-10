package config

type DatabaseKind string

const (
	DatabaseKindPostgreSQL DatabaseKind = "postgresql"
)

type Database struct {
	// Kind is the type of database used for monetr, currently only postgresql is
	// supported but in the future sqlite will be as well.
	Kind       DatabaseKind `yaml:"kind"`
	PostgreSQL *PostgreSQL  `yaml:"postgreSql"`
}

type PostgreSQL struct {
	Address            string `yaml:"address"`
	Port               int    `yaml:"port"`
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	Database           string `yaml:"database"`
	TLS                bool   `yaml:"tls"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	CACertificatePath  string `yaml:"caCertificatePath"`
	KeyPath            string `yaml:"keyPath"`
	CertificatePath    string `yaml:"certificatePath"`
	Migrate            bool   `yaml:"migrate"`
}
