resource "vault_mount" "plaid-client-secrets" {
  path = "secret/customers/plaid"
  type = "kv-v2"
  description = "KV store used for Plaid client credentials"
}