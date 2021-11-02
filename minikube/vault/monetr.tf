data "vault_policy_document" "monetr" {
  rule {
    path = "${vault_mount.plaid-client-secrets.path}/data/*"
    capabilities = [
      "create",
      "read",
      "update",
      "delete",
    ]
    description = "Allow monetr to manage client secrets."
  }
}

resource "vault_policy" "monetr" {
  name = "monetr"
  policy = data.vault_policy_document.monetr.hcl
}

resource "vault_kubernetes_auth_backend_role" "monetr" {
  backend = vault_auth_backend.kubernetes.path
  role_name = "monetr"
  bound_service_account_names = [
    "monetr",
  ]
  bound_service_account_namespaces = [
    "default",
  ]
  token_ttl = 3600
  token_policies = [
    vault_policy.monetr.name,
  ]
}
