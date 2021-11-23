terraform {
  required_providers {
    vault = {
      source = "hashicorp/vault"
      version = "3.0.1"
    }
  }
}

provider "vault" {
  address = "https://vault.monetr.mini"
}