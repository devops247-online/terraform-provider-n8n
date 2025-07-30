package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/devops247-online/terraform-provider-n8n/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &LDAPConfigResource{}
var _ resource.ResourceWithImportState = &LDAPConfigResource{}

func NewLDAPConfigResource() resource.Resource {
	return &LDAPConfigResource{}
}

// LDAPConfigResource defines the resource implementation.
type LDAPConfigResource struct {
	client *client.Client
}

// LDAPConfigResourceModel describes the resource data model.
type LDAPConfigResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	ServerURL              types.String `tfsdk:"server_url"`
	BindDN                 types.String `tfsdk:"bind_dn"`
	BindPassword           types.String `tfsdk:"bind_password"`
	SearchBase             types.String `tfsdk:"search_base"`
	SearchFilter           types.String `tfsdk:"search_filter"`
	UserIDAttribute        types.String `tfsdk:"user_id_attribute"`
	UserEmailAttribute     types.String `tfsdk:"user_email_attribute"`
	UserFirstNameAttribute types.String `tfsdk:"user_first_name_attribute"`
	UserLastNameAttribute  types.String `tfsdk:"user_last_name_attribute"`
	GroupSearchBase        types.String `tfsdk:"group_search_base"`
	GroupSearchFilter      types.String `tfsdk:"group_search_filter"`
	TLSEnabled             types.Bool   `tfsdk:"tls_enabled"`
	CACertificate          types.String `tfsdk:"ca_certificate"`
}

func (r *LDAPConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ldap_config"
}

func (r *LDAPConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages LDAP configuration for n8n Enterprise. This resource configures LDAP authentication and user synchronization.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "LDAP configuration identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_url": schema.StringAttribute{
				MarkdownDescription: "LDAP server URL (e.g., ldap://ldap.example.com:389 or ldaps://ldap.example.com:636)",
				Required:            true,
			},
			"bind_dn": schema.StringAttribute{
				MarkdownDescription: "Bind DN for LDAP connection (e.g., cn=admin,dc=example,dc=com)",
				Required:            true,
			},
			"bind_password": schema.StringAttribute{
				MarkdownDescription: "Bind password for LDAP connection",
				Required:            true,
				Sensitive:           true,
			},
			"search_base": schema.StringAttribute{
				MarkdownDescription: "User search base DN (e.g., ou=users,dc=example,dc=com)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"search_filter": schema.StringAttribute{
				MarkdownDescription: "User search filter (e.g., (uid={{username}}))",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("(uid={{username}})"),
			},
			"user_id_attribute": schema.StringAttribute{
				MarkdownDescription: "Attribute for user ID",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("uid"),
			},
			"user_email_attribute": schema.StringAttribute{
				MarkdownDescription: "Attribute for user email",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("mail"),
			},
			"user_first_name_attribute": schema.StringAttribute{
				MarkdownDescription: "Attribute for user first name",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("givenName"),
			},
			"user_last_name_attribute": schema.StringAttribute{
				MarkdownDescription: "Attribute for user last name",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("sn"),
			},
			"group_search_base": schema.StringAttribute{
				MarkdownDescription: "Group search base DN (e.g., ou=groups,dc=example,dc=com)",
				Optional:            true,
			},
			"group_search_filter": schema.StringAttribute{
				MarkdownDescription: "Group search filter (e.g., (member={{userDN}}))",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("(member={{userDN}})"),
			},
			"tls_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable TLS connection",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"ca_certificate": schema.StringAttribute{
				MarkdownDescription: "CA certificate for TLS connection (PEM format)",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *LDAPConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LDAPConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LDAPConfigResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create LDAP config object
	config := &client.LDAPConfig{
		ServerURL:              data.ServerURL.ValueString(),
		BindDN:                 data.BindDN.ValueString(),
		BindPassword:           data.BindPassword.ValueString(),
		SearchBase:             data.SearchBase.ValueString(),
		SearchFilter:           data.SearchFilter.ValueString(),
		UserIDAttribute:        data.UserIDAttribute.ValueString(),
		UserEmailAttribute:     data.UserEmailAttribute.ValueString(),
		UserFirstNameAttribute: data.UserFirstNameAttribute.ValueString(),
		UserLastNameAttribute:  data.UserLastNameAttribute.ValueString(),
		GroupSearchBase:        data.GroupSearchBase.ValueString(),
		GroupSearchFilter:      data.GroupSearchFilter.ValueString(),
		TLSEnabled:             data.TLSEnabled.ValueBool(),
		CACertificate:          data.CACertificate.ValueString(),
	}

	// Update LDAP config via API (LDAP config is a singleton, so we use update)
	updatedConfig, err := r.client.UpdateLDAPConfig(config)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create LDAP config, got error: %s", err))
		return
	}

	// Update model with response data
	r.updateModelFromLDAPConfig(&data, updatedConfig)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LDAPConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LDAPConfigResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get LDAP config from API
	config, err := r.client.GetLDAPConfig()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read LDAP config, got error: %s", err))
		return
	}

	// Update model with response data
	r.updateModelFromLDAPConfig(&data, config)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LDAPConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LDAPConfigResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create LDAP config object for update
	config := &client.LDAPConfig{
		ServerURL:              data.ServerURL.ValueString(),
		BindDN:                 data.BindDN.ValueString(),
		BindPassword:           data.BindPassword.ValueString(),
		SearchBase:             data.SearchBase.ValueString(),
		SearchFilter:           data.SearchFilter.ValueString(),
		UserIDAttribute:        data.UserIDAttribute.ValueString(),
		UserEmailAttribute:     data.UserEmailAttribute.ValueString(),
		UserFirstNameAttribute: data.UserFirstNameAttribute.ValueString(),
		UserLastNameAttribute:  data.UserLastNameAttribute.ValueString(),
		GroupSearchBase:        data.GroupSearchBase.ValueString(),
		GroupSearchFilter:      data.GroupSearchFilter.ValueString(),
		TLSEnabled:             data.TLSEnabled.ValueBool(),
		CACertificate:          data.CACertificate.ValueString(),
	}

	// Update LDAP config via API
	updatedConfig, err := r.client.UpdateLDAPConfig(config)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update LDAP config, got error: %s", err))
		return
	}

	// Update model with response data
	r.updateModelFromLDAPConfig(&data, updatedConfig)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LDAPConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// LDAP config cannot be deleted, only disabled
	// We could potentially clear the configuration by setting minimal values
	resp.Diagnostics.AddWarning(
		"LDAP Configuration Not Deleted",
		"LDAP configuration cannot be deleted from n8n. The resource has been removed from Terraform state, but the LDAP configuration remains in n8n. To disable LDAP, update the configuration with appropriate values.",
	)
}

func (r *LDAPConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// LDAP config is a singleton, so we use a fixed ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "ldap")...)
}

// Helper function to update model from API response
func (r *LDAPConfigResource) updateModelFromLDAPConfig(model *LDAPConfigResourceModel, config *client.LDAPConfig) {
	model.ID = types.StringValue("ldap") // LDAP config is a singleton
	model.ServerURL = types.StringValue(config.ServerURL)
	model.BindDN = types.StringValue(config.BindDN)
	// Don't update bind_password from response for security
	model.SearchBase = types.StringValue(config.SearchBase)
	model.SearchFilter = types.StringValue(config.SearchFilter)
	model.UserIDAttribute = types.StringValue(config.UserIDAttribute)
	model.UserEmailAttribute = types.StringValue(config.UserEmailAttribute)
	model.UserFirstNameAttribute = types.StringValue(config.UserFirstNameAttribute)
	model.UserLastNameAttribute = types.StringValue(config.UserLastNameAttribute)
	model.GroupSearchBase = types.StringValue(config.GroupSearchBase)
	model.GroupSearchFilter = types.StringValue(config.GroupSearchFilter)
	model.TLSEnabled = types.BoolValue(config.TLSEnabled)
	// Don't update ca_certificate from response for security
}
