package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &MongoDBProvider{}
)

// MongoDBProvider defines the provider implementation
type MongoDBProvider struct{}

// New creates a new provider
func New() provider.Provider {
	return &MongoDBProvider{}
}

// Metadata returns provider metadata
func (p *MongoDBProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mongodocs"
}

// Schema defines the provider configuration schema
func (p *MongoDBProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with MongoDB and AWS Secrets Manager",
		Attributes: map[string]schema.Attribute{
			"connection_uri": schema.StringAttribute{
				Optional:    true,
				Description: "Default MongoDB connection URI",
				// Default:     "mongodb://localhost:27017",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Default MongoDB username",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Default MongoDB password",
			},
		},
	}
}

// Configure prepares a MongoDB client
func (p *MongoDBProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// No configuration needed for this example
}

// Resources returns available resources
func (p *MongoDBProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMongoDocumentResource,
	}
}

// DataSources returns available data sources
func (p *MongoDBProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
