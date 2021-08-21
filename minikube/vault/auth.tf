resource "vault_auth_backend" "kubernetes" {
  type = "kubernetes"
}

resource "vault_kubernetes_auth_backend_config" "example" {
  backend = vault_auth_backend.kubernetes.path
  kubernetes_host = var.kubernetes_address
  kubernetes_ca_cert = base64decode(var.kubernetes_certificate_b64)
  token_reviewer_jwt = var.kubernetes_reviewer_jwt
  issuer = "kubernetes.io/serviceaccount"
  disable_iss_validation = true
}

resource "vault_auth_backend" "userpass" {
  type = "userpass"
}

resource "vault_generic_endpoint" "monetr-user" {
  depends_on           = [vault_auth_backend.userpass]
  path                 = "auth/userpass/users/monetr"
  ignore_absent_fields = true

  data_json = jsonencode({
    policies = [
      vault_policy.rest-api-service-policy.name,
    ]
    password = "password"
  })
}
