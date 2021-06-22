resource "vault_mount" "plaid-client-secrets" {
  path = "plaid/clients"
  type = "kv-v2"
  description = "KV store used for Plaid client credentials"
}