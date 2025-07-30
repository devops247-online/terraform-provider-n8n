package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/devops247-online/terraform-provider-n8n/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CredentialResource{}
var _ resource.ResourceWithImportState = &CredentialResource{}

func NewCredentialResource() resource.Resource {
	return &CredentialResource{}
}

// CredentialResource defines the resource implementation.
type CredentialResource struct {
	client *client.Client
}

// CredentialResourceModel describes the resource data model.
type CredentialResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Data       types.String `tfsdk:"data"`
	NodeAccess types.List   `tfsdk:"node_access"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

// Supported credential types for validation
var supportedCredentialTypes = []string{
	"httpBasicAuth",
	"apiKey",
	"oAuth2Api",
	"oAuth1Api",
	"googleOAuth2Api",
	"awsApi",
	"azureApi",
	"microsoftGraphApi",
	"httpDigestAuth",
	"httpHeaderAuth",
	"httpQueryAuth",
	"jwtAuth",
	"bearerTokenAuth",
	"samlAuth",
	"ldapAuth",
	"slackOAuth2Api",
	"githubOAuth2Api",
	"gitlabOAuth2Api",
	"discordOAuth2Api",
	"shopifyOAuth2Api",
	"stripeApi",
	"twilioApi",
	"sendGridApi",
	"mailgunApi",
	"dropboxOAuth2Api",
	"googleDriveOAuth2Api",
	"onedriveOAuth2Api",
	"boxOAuth2Api",
	"salesforceOAuth2Api",
	"hubspotOAuth2Api",
	"zendeskOAuth2Api",
	"jiraCloudApi",
	"confluenceCloudApi",
	"atlassianOAuth2Api",
	"trelloApi",
	"asanaOAuth2Api",
	"mondayComOAuth2Api",
	"notionOAuth2Api",
	"airtableOAuth2Api",
	"clickUpOAuth2Api",
	"linearOAuth2Api",
	"figmaOAuth2Api",
	"canvaOAuth2Api",
	"youtubeOAuth2Api",
	"spotifyOAuth2Api",
	"twitterOAuth2Api",
	"facebookGraphApi",
	"instagramGraphApi",
	"linkedInOAuth2Api",
	"telegramBotApi",
	"whatsappBusinessApi",
	"openAiApi",
	"anthropicApi",
	"huggingFaceApi",
	"cohereApi",
	"mistralApi",
	"groqApi",
}

func (r *CredentialResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

func (r *CredentialResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an n8n credential securely. Credentials store authentication information for services and APIs used by workflows, with proper handling of sensitive data.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Credential identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the credential. Must be unique within the n8n instance.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of credential (e.g., 'httpBasicAuth', 'oAuth2Api', 'apiKey'). Determines the required data fields.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data": schema.StringAttribute{
				MarkdownDescription: "JSON string containing the credential configuration data. This field is sensitive and will be encrypted in state.",
				Optional:            true,
				Sensitive:           true,
			},
			"node_access": schema.ListAttribute{
				MarkdownDescription: "List of node names that can access this credential. If empty, all nodes can access it.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the credential was created",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the credential was last updated",
				Computed:            true,
			},
		},
	}
}

func (r *CredentialResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *CredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CredentialResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate credential type
	if err := r.validateCredentialType(data.Type.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("type"),
			"Invalid Credential Type",
			err.Error(),
		)
		return
	}

	// Create credential object
	credential := &client.Credential{
		Name: data.Name.ValueString(),
		Type: data.Type.ValueString(),
	}

	// Parse and validate credential data if provided
	if !data.Data.IsNull() && data.Data.ValueString() != "" {
		var credData map[string]interface{}
		if err := json.Unmarshal([]byte(data.Data.ValueString()), &credData); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("data"),
				"Invalid JSON",
				fmt.Sprintf("Unable to parse credential data JSON: %s", err),
			)
			return
		}

		// Validate credential data based on type
		if err := r.validateCredentialData(data.Type.ValueString(), credData); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("data"),
				"Invalid Credential Data",
				err.Error(),
			)
			return
		}

		credential.Data = credData
	}

	// Handle node access
	if !data.NodeAccess.IsNull() {
		var nodeAccess []string
		resp.Diagnostics.Append(data.NodeAccess.ElementsAs(ctx, &nodeAccess, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		credential.SharedWith = nodeAccess
	}

	// Create credential via API
	createdCredential, err := r.client.CreateCredential(credential)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create credential, got error: %s", err))
		return
	}

	// Update model with response data
	r.updateModelFromCredential(&data, createdCredential)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CredentialResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get credential from API
	credential, err := r.client.GetCredential(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read credential, got error: %s", err))
		return
	}

	// Update model with response data
	r.updateModelFromCredential(&data, credential)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CredentialResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create credential object for update
	credential := &client.Credential{
		Name: data.Name.ValueString(),
		Type: data.Type.ValueString(),
	}

	// Parse and validate credential data if provided
	if !data.Data.IsNull() && data.Data.ValueString() != "" {
		var credData map[string]interface{}
		if err := json.Unmarshal([]byte(data.Data.ValueString()), &credData); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("data"),
				"Invalid JSON",
				fmt.Sprintf("Unable to parse credential data JSON: %s", err),
			)
			return
		}

		// Validate credential data based on type
		if err := r.validateCredentialData(data.Type.ValueString(), credData); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("data"),
				"Invalid Credential Data",
				err.Error(),
			)
			return
		}

		credential.Data = credData
	}

	// Handle node access
	if !data.NodeAccess.IsNull() {
		var nodeAccess []string
		resp.Diagnostics.Append(data.NodeAccess.ElementsAs(ctx, &nodeAccess, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		credential.SharedWith = nodeAccess
	}

	// Update credential via API
	updatedCredential, err := r.client.UpdateCredential(data.ID.ValueString(), credential)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update credential, got error: %s", err))
		return
	}

	// Update model with response data
	r.updateModelFromCredential(&data, updatedCredential)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CredentialResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete credential via API
	err := r.client.DeleteCredential(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete credential, got error: %s", err))
		return
	}
}

func (r *CredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// validateCredentialType validates that the credential type is supported
func (r *CredentialResource) validateCredentialType(credType string) error {
	if credType == "" {
		return fmt.Errorf("credential type is required")
	}

	if !slices.Contains(supportedCredentialTypes, credType) {
		return fmt.Errorf("unsupported credential type: %s. Supported types: %s", credType, strings.Join(supportedCredentialTypes, ", "))
	}

	return nil
}

// validateCredentialData validates the credential data based on type
func (r *CredentialResource) validateCredentialData(credType string, data map[string]interface{}) error {
	if data == nil {
		return nil
	}

	// Type-specific validation
	switch credType {
	case "httpBasicAuth":
		if _, hasUser := data["user"]; !hasUser {
			return fmt.Errorf("httpBasicAuth credential requires 'user' field")
		}
		if _, hasPassword := data["password"]; !hasPassword {
			return fmt.Errorf("httpBasicAuth credential requires 'password' field")
		}

	case "apiKey":
		if _, hasApiKey := data["apiKey"]; !hasApiKey {
			return fmt.Errorf("apiKey credential requires 'apiKey' field")
		}

	case "oAuth2Api":
		if _, hasClientId := data["clientId"]; !hasClientId {
			return fmt.Errorf("oAuth2Api credential requires 'clientId' field")
		}
		if _, hasClientSecret := data["clientSecret"]; !hasClientSecret {
			return fmt.Errorf("oAuth2Api credential requires 'clientSecret' field")
		}

	case "bearerTokenAuth":
		if _, hasToken := data["token"]; !hasToken {
			return fmt.Errorf("bearerTokenAuth credential requires 'token' field")
		}

	case "httpHeaderAuth":
		if _, hasName := data["name"]; !hasName {
			return fmt.Errorf("httpHeaderAuth credential requires 'name' field")
		}
		if _, hasValue := data["value"]; !hasValue {
			return fmt.Errorf("httpHeaderAuth credential requires 'value' field")
		}

	case "awsApi":
		if _, hasAccessKeyId := data["accessKeyId"]; !hasAccessKeyId {
			return fmt.Errorf("awsApi credential requires 'accessKeyId' field")
		}
		if _, hasSecretAccessKey := data["secretAccessKey"]; !hasSecretAccessKey {
			return fmt.Errorf("awsApi credential requires 'secretAccessKey' field")
		}

	case "googleOAuth2Api":
		if _, hasClientId := data["clientId"]; !hasClientId {
			return fmt.Errorf("googleOAuth2Api credential requires 'clientId' field")
		}
		if _, hasClientSecret := data["clientSecret"]; !hasClientSecret {
			return fmt.Errorf("googleOAuth2Api credential requires 'clientSecret' field")
		}
	}

	return nil
}

// Helper function to update model from API response
func (r *CredentialResource) updateModelFromCredential(model *CredentialResourceModel, credential *client.Credential) {
	model.ID = types.StringValue(credential.ID)
	model.Name = types.StringValue(credential.Name)
	model.Type = types.StringValue(credential.Type)

	// Convert credential data to JSON string (but keep it sensitive)
	// Note: We don't include sensitive data in read operations for security
	if len(credential.Data) > 0 {
		// Only update data field if it's currently set (to preserve sensitive data in state)
		if !model.Data.IsNull() {
			if dataJSON, err := json.Marshal(credential.Data); err == nil {
				model.Data = types.StringValue(string(dataJSON))
			}
		}
	}

	// Handle node access / shared with
	if credential.SharedWith != nil {
		nodeAccessValues := make([]attr.Value, len(credential.SharedWith))
		for i, node := range credential.SharedWith {
			nodeAccessValues[i] = types.StringValue(node)
		}
		model.NodeAccess = types.ListValueMust(types.StringType, nodeAccessValues)
	} else {
		model.NodeAccess = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if credential.CreatedAt != nil {
		model.CreatedAt = types.StringValue(credential.CreatedAt.Format("2006-01-02T15:04:05Z"))
	}

	if credential.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(credential.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	}
}
