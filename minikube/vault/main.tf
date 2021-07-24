terraform {
  required_providers {
    vault = {
      source = "hashicorp/vault"
      version = "2.22.1"
    }
  }
}

provider "vault" {
  address = "https://vault.monetr.mini"
}