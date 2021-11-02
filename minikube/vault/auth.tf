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

// monetr-user is just a basic user-password authentication for the API. This is to be used when trying to debug
// permissions or trying to dignose issues outside of Kubernetes. As it is easier to provision access using a
// username and password than it is to try to use the Kubernetes authentication, outside kube.
// The user is "monetr", the password is "password".
resource "vault_generic_endpoint" "monetr-user" {
  depends_on           = [vault_auth_backend.userpass]
  path                 = "auth/userpass/users/monetr"
  ignore_absent_fields = true

  data_json = jsonencode({
    policies = [
      vault_policy.monetr.name,
    ]
    password = "password"
  })
}
