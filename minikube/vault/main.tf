terraform {
  required_providers {
    vault = {
      source = "hashicorp/vault"
      version = "3.3.1"
    }
  }
}

provider "vault" {
  address = "https://vault.monetr.mini"
}