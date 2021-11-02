resource "vault_mount" "plaid-client-secrets" {
  path        = "customers/plaid"
  type        = "kv-v2"
  description = "KV store used for Plaid client credentials"
}

resource "vault_mount" "database" {
  path        = "monetr/database"
  type        = "database"
  description = "Vault provisioning access to PostgreSQL"
}

resource "vault_database_secret_backend_connection" "postgres" {
  backend       = vault_mount.database.path
  name          = "postgres"
  allowed_roles = [
    vault_kubernetes_auth_backend_role.monetr.role_name,
  ]

  postgresql {
    connection_url = "postgres://postgres@postgres.default.svc.cluster.local:5432/postgres?sslmode=disable"
  }
}

resource "vault_database_secret_backend_role" "monetr" {
  backend             = vault_mount.database.path
  name                = "monetr"
  db_name             = vault_database_secret_backend_connection.postgres.name
  creation_statements = ["CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';"]
}