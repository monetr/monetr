data "vault_policy_document" "rest-api-policy" {
  rule {
    path = "${vault_mount.plaid-client-secrets.path}/data/*"
    capabilities = [
      "create",
      "read",
      "update",
      "delete",
    ]
    description = "Allow the REST API to manage client secrets."
  }
}

resource "vault_policy" "rest-api-service-policy" {
  name = "rest-api-service-policy"
  policy = data.vault_policy_document.rest-api-policy.hcl
}

resource "vault_kubernetes_auth_backend_role" "rest-api" {
  backend = vault_auth_backend.kubernetes.path
  role_name = "rest-api"
  bound_service_account_names = [
    "rest-api",
  ]
  bound_service_account_namespaces = [
    "*",
  ]
  token_ttl = 3600
  token_policies = [
    vault_policy.rest-api-service-policy.name,
  ]
}
