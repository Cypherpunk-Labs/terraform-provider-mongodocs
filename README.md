# terraform-provider-mongodocs
Terraform Provider for adding documents on MongoDB



Terraform Cloud Custom Go
To include a custom provider with Terraform Cloud, you need to compile the custom provider and include it directly in the project directory. This involves creating a .terraformignore file in your project directory and putting in the necessary configuration. For example, you can include the following in your .terraformignore file:

.terraform
.git
.gtm
*.tfstate
This ensures that local plugins do not generate some tfstate for mapping the local plugin directory, which can cause conflicts when uploading to Terraform Cloud.

You also need to ensure that the custom provider is compiled for the correct platform. For Terraform Cloud, you should target the platform for Go in the build step, as the Terraform Cloud environment expects Linux and amd64 as the target. This can be done using commands like goreleaser build --snapshot to generate all binaries for you.

The custom provider can be included in the project directory under the terraform.d/plugins/linux_amd64 path. This allows Terraform Cloud to run the custom provider and return the plan successfully.

For more detailed steps and troubleshooting, you can refer to the documentation on compiling a custom provider and including it for Terraform Cloud.

```
# Configure the MongoDB provider
terraform {
  required_providers {
    mongodocs = {
      source = "your-namespace/mongodb"
    }
  }
}

# Optional provider configuration
provider "mongodocs" {
  connection_uri = "mongodb://localhost:27017"
  # username = "optional_username"
  # password = "optional_password"
}

# Create a MongoDB document using direct content
resource "mongodocs_document" "example_doc" {
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
  database    = "testdb"
  collection  = "secrets"
  secret_name = "my-document-secret"
}
```

https://developer.hashicorp.com/terraform/registry/providers/publishing