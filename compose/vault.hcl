storage "file" {
    path = "/vault/data"
}

ui = true

seal "awskms" {
  region     = "us-east-1"
  access_key = "foo"
  secret_key = "bar"
  kms_key_id = "bc436485-5092-42b8-92a3-0aa8b93536dc"
  endpoint   = "http://kms:8080"
}

listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_disable = "true"
}

api_addr = "http://0.0.0.0:8200"
cluster_addr = "https://0.0.0.0:8201"

log_requests_level = "trace"
