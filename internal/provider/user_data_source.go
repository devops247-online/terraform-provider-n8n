package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/devops247-online/terraform-provider-n8n/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

// UserDataSource defines the data source implementation.
type UserDataSource struct {
	client *client.Client
}

// UserDataSourceModel describes the data source data model.
type UserDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Role      types.String `tfsdk:"role"`
	IsOwner   types.Bool   `tfsdk:"is_owner"`
	IsPending types.Bool   `tfsdk:"is_pending"`
	Settings  types.Object `tfsdk:"settings"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest,
	resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about an n8n user. You can look up a user by their ID or email address.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "User identifier. Either id or email must be provided.",
				Optional:            true,
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "User email address. Either id or email must be provided.",
				Optional:            true,
				Computed:            true,
			},
			"first_name": schema.StringAttribute{
				MarkdownDescription: "User's first name",
				Computed:            true,
			},
			"last_name": schema.StringAttribute{
				MarkdownDescription: "User's last name",
				Computed:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "User role (e.g., 'admin', 'member', 'editor')",
				Computed:            true,
			},
			"is_owner": schema.BoolAttribute{
				MarkdownDescription: "Whether the user is an owner of the n8n instance",
				Computed:            true,
			},
			"is_pending": schema.BoolAttribute{
				MarkdownDescription: "Whether the user invitation is pending",
				Computed:            true,
			},
			"settings": schema.SingleNestedAttribute{
				MarkdownDescription: "User-specific settings",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"theme": schema.StringAttribute{
						MarkdownDescription: "User's preferred theme (e.g., 'light', 'dark')",
						Computed:            true,
					},
					"allow_sso_manual_login": schema.BoolAttribute{
						MarkdownDescription: "Whether to allow SSO manual login for this user",
						Computed:            true,
					},
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the user was created",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the user was last updated",
				Computed:            true,
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either ID or email is provided
	if data.ID.IsNull() && data.Email.IsNull() {
		resp.Diagnostics.AddError(
			"Missing User Identifier",
			"Either 'id' or 'email' must be provided to look up a user.",
		)
		return
	}

	var user *client.User
	var err error

	// Look up user by ID if provided, otherwise by email
	if !data.ID.IsNull() {
		user, err = d.client.GetUser(data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user by ID, got error: %s", err))
			return
		}
	} else {
		// Look up user by email - we need to list users and find the one with matching email
		users, err := d.client.GetUsers(nil)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list users, got error: %s", err))
			return
		}

		emailToFind := data.Email.ValueString()
		var foundUser *client.User
		for _, u := range users.Data {
			if u.Email == emailToFind {
				foundUser = &u
				break
			}
		}

		if foundUser == nil {
			resp.Diagnostics.AddError("User Not Found", fmt.Sprintf("No user found with email: %s", emailToFind))
			return
		}
		user = foundUser
	}

	// Update model with user data
	d.updateModelFromUser(&data, user)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to update model from API response
func (d *UserDataSource) updateModelFromUser(model *UserDataSourceModel, user *client.User) {
	model.ID = types.StringValue(user.ID)
	model.Email = types.StringValue(user.Email)

	if user.FirstName != "" {
		model.FirstName = types.StringValue(user.FirstName)
	}

	if user.LastName != "" {
		model.LastName = types.StringValue(user.LastName)
	}

	if user.Role != "" {
		model.Role = types.StringValue(user.Role)
	}

	model.IsOwner = types.BoolValue(user.IsOwner)
	model.IsPending = types.BoolValue(user.IsPending)

	// Handle settings
	if user.Settings.Theme != "" || user.Settings.AllowSSOManualLogin {
		settingsAttrs := map[string]attr.Value{
			"theme":                  types.StringValue(user.Settings.Theme),
			"allow_sso_manual_login": types.BoolValue(user.Settings.AllowSSOManualLogin),
		}
		model.Settings = types.ObjectValueMust(
			map[string]attr.Type{
				"theme":                  types.StringType,
				"allow_sso_manual_login": types.BoolType,
			},
			settingsAttrs,
		)
	} else {
		// Create empty settings object
		settingsAttrs := map[string]attr.Value{
			"theme":                  types.StringNull(),
			"allow_sso_manual_login": types.BoolNull(),
		}
		model.Settings = types.ObjectValueMust(
			map[string]attr.Type{
				"theme":                  types.StringType,
				"allow_sso_manual_login": types.BoolType,
			},
			settingsAttrs,
		)
	}

	if user.CreatedAt != nil {
		model.CreatedAt = types.StringValue(user.CreatedAt.Format("2006-01-02T15:04:05Z"))
	}

	if user.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(user.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	}
}
