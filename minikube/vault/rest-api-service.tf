data "vault_policy_document" "rest-api-policy" {
  // There are two rules here because I still have no idea what I'm doing when it comes to trying to provision this
  // stuff. I'm getting permissions errors for both paths so this is likely to change in the future. But again this
  // is only used for local development against vault.
  rule {
    path = "${vault_mount.plaid-client-secrets.path}/*"
    capabilities = [
      "create",
      "read",
      "update",
      "delete",
    ]
    description = "Allow the REST API to manage client secrets."
  }

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
