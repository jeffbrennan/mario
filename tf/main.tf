# Configure the Azure provider
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0.2"
    }
    databricks = {
      source = "databricks/databricks"
    }
  }

  required_version = ">= 1.1.0"
}

provider "azurerm" {
  features {}
}


locals {
  prefix = "jb"
}

resource "azurerm_resource_group" "rg" {
  name     = var.resource_group_name
  location = "eastus2"
}


resource "azurerm_data_factory" "adf" {
  name                   = var.data_factory_name
  location               = azurerm_resource_group.rg.location
  resource_group_name    = azurerm_resource_group.rg.name
  public_network_enabled = true
  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_storage_account" "storage" {
  name                     = var.storage_account_name
  resource_group_name      = azurerm_resource_group.rg.name
  location                 = azurerm_resource_group.rg.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_storage_container" "container" {
  name                  = var.container_name
  storage_account_name  = azurerm_storage_account.storage.name
  container_access_type = "private"
}

resource "azurerm_role_assignment" "blob_contributor" {
  scope                = azurerm_storage_account.storage.id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = var.user_object_id
}


resource "azurerm_databricks_workspace" "databricks" {
  name                        = "${local.prefix}-workspace"
  resource_group_name         = azurerm_resource_group.rg.name
  location                    = azurerm_resource_group.rg.location
  sku                         = "premium"
  managed_resource_group_name = "${local.prefix}-databricks-rg"
}

data "azurerm_client_config" "current" {}

resource "azurerm_key_vault" "vault" {
  name                       = "jb-vault"
  location                   = azurerm_resource_group.rg.location
  resource_group_name        = azurerm_resource_group.rg.name
  tenant_id                  = data.azurerm_client_config.current.tenant_id
  sku_name                   = "standard"
  soft_delete_retention_days = 7
}