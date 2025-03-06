When using Secret. 

Plan: 1 to add, 0 to change, 0 to destroy.

Do you want to perform these actions in workspace "backend-data-layer"?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

mongodocs_document.secret_doc: Creating...
╷
│ Error: Provider produced inconsistent result after apply
│ 
│ When applying changes to mongodocs_document.secret_doc, provider
│ "provider[\"registry.terraform.io/cypherpunk-labs/mongodocs\"]" produced an
│ unexpected new value: .content: was null, but now cty.StringVal("{
│ \"primaryLayerModel\": \"hosted-llm-1\" }\n").
│ 
│ This is a bug in the provider, which should be reported in the provider's
│ own issue tracker.
╵

the content input is null, but in the create func, this is being written to in the state.
decided to separate content and doccontent in the schema.