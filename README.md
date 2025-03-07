# terraform-provider-mongodocs
Terraform Provider for adding documents on MongoDB

The purpose of this provider is to give the capability to create documents in mongodb where in the official atlas provider does not provide this capiability.
databases and collection that do not exist will be created on the fly, and a unique document id will be generated by mongodb and used to track the resource in terraform.
This provider is used to retrieve a secret from AWS to use as a document, this could be config for an application that contains sensitive secrets. 
You can use either 'content' or 'secret_name' but not both, secret_name will take preference. 
'secret_name' needs to be a valid AWS Secret name. 
'content' needs to be a valid json document (do not include an ID at the root level if you copied from another document).
We elected to use the credentials and connection parameters at the resource level so that we can for_each through our clusters. 

[Terraform Provider Docs](docs/index.md)

```
# Configure the MongoDB provider
terraform {
  required_providers {
    mongodocs = {
      source = "Cypherpunk-Labs/mongodocs"
    }
  }
}

# Optional provider configuration
provider "mongodocs" {}

# Create a MongoDB document using direct content
resource "mongodocs_document" "example_doc" {
  connection_uri = "mongodb://localhost:27017"
  # username = "optional_username"
  # password = "optional_password"
  database    = "testdb"
  collection  = "documents"
  content     = jsonencode({
    title       = "Example Document"
    description = "This is a sample document created via Terraform"
    created_at  = timestamp()
  })
}

# Create a MongoDB document using AWS Secrets Manager
resource "mongodocs_document" "secret_doc" {
  connection_uri = "mongodb://localhost:27017"
  # username = "optional_username"
  # password = "optional_password"
  database    = "testdb"
  collection  = "secrets"
  secret_name = "my-document-secret"
}
```

# index

## generating plugin docs

execute 'tfplugindocs' which is installed as per https://github.com/hashicorp/terraform-plugin-docs

## Useful links for developing a provider 

- https://developer.hashicorp.com/terraform/registry/providers/publishing
- https://developer.hashicorp.com/terraform/plugin/debugging