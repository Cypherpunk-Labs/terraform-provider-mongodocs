package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDocumentResourceModel describes the resource data model
type MongoDocumentResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ConnectionURI types.String `tfsdk:"connection_uri"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	Database      types.String `tfsdk:"database"`
	Collection    types.String `tfsdk:"collection"`
	Content       types.String `tfsdk:"content"`
	DocContent    types.String `tfsdk:"doc_content"`
	SecretName    types.String `tfsdk:"secret_name"`
}

// NewMongoDocumentResource is a helper function to create the resource
func NewMongoDocumentResource() resource.Resource {
	return &MongoDocumentResource{}
}

// MongoDocumentResource is the resource implementation
type MongoDocumentResource struct{}

// Metadata returns the resource type name
func (r *MongoDocumentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_document"
}

// Schema defines the schema for the resource
func (r *MongoDocumentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a MongoDB document",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the document",
			},
			"connection_uri": schema.StringAttribute{
				Required:    true,
				Description: "MongoDB connection URI",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "MongoDB username",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "MongoDB password",
			},
			"database": schema.StringAttribute{
				Required:    true,
				Description: "Name of the MongoDB database",
			},
			"collection": schema.StringAttribute{
				Required:    true,
				Description: "Name of the MongoDB collection",
			},
			"content": schema.StringAttribute{
				Optional:    true,
				Description: "JSON content of the document",
			},
			"doc_content": schema.StringAttribute{
				Optional:    true,
				Description: "JSON content of the document",
			},
			"secret_name": schema.StringAttribute{
				Optional:    true,
				Description: "AWS Secrets Manager secret name for document content",
			},
		},
	}
}

// Create creates a new resource
func (r *MongoDocumentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MongoDocumentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine document content
	var docContent string
	if !plan.SecretName.IsNull() && plan.SecretName.ValueString() != "" {
		// Fetch content from AWS Secrets Manager
		secretContent, err := fetchDocumentFromAWSSecretsManager(ctx, plan.SecretName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Fetch Secret",
				fmt.Sprintf("Unable to retrieve document from AWS Secrets Manager: %v", err),
			)
			return
		}
		docContent = secretContent
	} else {
		if plan.Content.IsNull() {
			resp.Diagnostics.AddError(
				"Missing Document Content",
				"Either 'content' or 'secret_name' must be provided",
			)
			return
		} else {
			docContent = plan.Content.ValueString()
		}
	}

	// Connect to MongoDB
	client, err := r.connectToMongoDB(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to connect to MongoDB",
			fmt.Sprintf("Unable to connect to MongoDB: %v", err),
		)
		return
	}
	defer client.Disconnect(ctx)

	// Prepare document
	var doc interface{}
	err = json.Unmarshal([]byte(docContent), &doc)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Document Content",
			fmt.Sprintf("Unable to parse document content: %v", err),
		)
		return
	}

	// Insert document
	collection := client.Database(plan.Database.ValueString()).Collection(plan.Collection.ValueString())
	result, err := collection.InsertOne(ctx, doc)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Insert Document",
			fmt.Sprintf("Unable to insert document: %v", err),
		)
		return
	}

	// After successful insertion
	plan.ID = types.StringValue(result.InsertedID.(primitive.ObjectID).Hex())

	// Set ID and content in state
	// plan.ID = types.StringValue(fmt.Sprintf("%v", result.InsertedID))
	// plan.DocContent = types.StringValue(docContent)
	if !plan.DocContent.IsNull() || !plan.Content.IsNull() {
		plan.DocContent = types.StringValue(docContent)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the resource state
func (r *MongoDocumentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MongoDocumentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Connect to MongoDB
	client, err := r.connectToMongoDB(ctx, state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to connect to MongoDB",
			fmt.Sprintf("Unable to connect to MongoDB: %v", err),
		)
		return
	}
	defer client.Disconnect(ctx)

	// Fetch document
	collection := client.Database(state.Database.ValueString()).Collection(state.Collection.ValueString())
	var result bson.M
	objectID, err := primitive.ObjectIDFromHex(state.ID.ValueString())
	if err != nil {
		// Handle error
		resp.Diagnostics.AddError(
			"Failed to read ObjectID from state",
			fmt.Sprintf("Unable to read objectID state: %v", err),
		)
		return
	}
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&result)
	// err = collection.FindOne(ctx, bson.M{"_id": state.ID.ValueString()}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Document deleted, remove from state
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Fetch Document",
			fmt.Sprintf("Unable to find document: %v", err),
		)
		return
	}

	// Convert result back to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Convert Document",
			fmt.Sprintf("Unable to convert document to JSON: %v", err),
		)
		return
	}
	state.DocContent = types.StringValue(string(jsonBytes))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates an existing resource
func (r *MongoDocumentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state MongoDocumentResourceModel
	req.Plan.Get(ctx, &plan)
	req.State.Get(ctx, &state)

	// Connect to MongoDB
	client, err := r.connectToMongoDB(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to connect to MongoDB",
			fmt.Sprintf("Unable to connect to MongoDB: %v", err),
		)
		return
	}
	defer client.Disconnect(ctx)

	//TODO: missing steps for secrethandling
	// Prepare updated document
	var doc interface{}
	err = json.Unmarshal([]byte(plan.DocContent.ValueString()), &doc)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Document Content",
			fmt.Sprintf("Unable to parse document content: %v", err),
		)
		return
	}

	// Update document
	collection := client.Database(plan.Database.ValueString()).Collection(plan.Collection.ValueString())
	objectID, err := primitive.ObjectIDFromHex(state.ID.ValueString())
	if err != nil {
		// Handle error
		resp.Diagnostics.AddError(
			"Failed to read ObjectID from state",
			fmt.Sprintf("Unable to read objectID state: %v", err),
		)
		return
	}
	_, err = collection.ReplaceOne(ctx, bson.M{"_id": objectID}, doc)
	// _, err = collection.ReplaceOne(ctx, bson.M{"_id": state.ID.ValueString()}, doc)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Update Document",
			fmt.Sprintf("Unable to update document: %v", err),
		)
		return
	}

	// Update state
	plan.ID = state.ID
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete removes the resource
func (r *MongoDocumentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MongoDocumentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Connect to MongoDB
	client, err := r.connectToMongoDB(ctx, state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to connect to MongoDB",
			fmt.Sprintf("Unable to connect to MongoDB: %v", err),
		)
		return
	}
	defer client.Disconnect(ctx)

	// Delete document
	collection := client.Database(state.Database.ValueString()).Collection(state.Collection.ValueString())
	objectID, err := primitive.ObjectIDFromHex(state.ID.ValueString())
	if err != nil {
		// Handle error
		resp.Diagnostics.AddError(
			"Failed to read ObjectID from state",
			fmt.Sprintf("Unable to read objectID state: %v", err),
		)
		return
	}
	// _, err = collection.DeleteOne(ctx, bson.M{"_id": state.ID.ValueString()})
	_, err = collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete Document",
			fmt.Sprintf("Unable to delete document: %v", err),
		)
		return
	}
}

// Helper method to connect to MongoDB
func (r *MongoDocumentResource) connectToMongoDB(ctx context.Context, model MongoDocumentResourceModel) (*mongo.Client, error) {
	// Prepare client options
	opts := options.Client().ApplyURI(model.ConnectionURI.ValueString())

	// Set authentication if credentials are provided
	if !model.Username.IsNull() && !model.Password.IsNull() {
		opts.SetAuth(options.Credential{
			Username: model.Username.ValueString(),
			Password: model.Password.ValueString(),
		})
	}

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to construct MongoDB client: %v", err)
	}

	// Ping to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	return client, nil
}

// fetchDocumentFromAWSSecretsManager retrieves document content from AWS Secrets Manager
func fetchDocumentFromAWSSecretsManager(ctx context.Context, secretName string) (string, error) {
	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS SDK config: %v", err)
	}

	// Create Secrets Manager client
	client := secretsmanager.NewFromConfig(cfg)

	// Retrieve the secret
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}

	result, err := client.GetSecretValue(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret: %v", err)
	}

	// Return the secret string
	return *result.SecretString, nil
}
