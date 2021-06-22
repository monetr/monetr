resource "vault_audit" "debugging" {
  type = "file"
  path = "stdout"
  local = false

  options = {
    file_path = "stdout"
    low_raw = true
    format = "json"
  }
}