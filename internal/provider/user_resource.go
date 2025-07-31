package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/devops247-online/terraform-provider-n8n/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

// UserResource defines the resource implementation.
type UserResource struct {
	client *client.Client
}

// UserResourceModel describes the resource data model.
type UserResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Role      types.String `tfsdk:"role"`
	Password  types.String `tfsdk:"password"`
	IsOwner   types.Bool   `tfsdk:"is_owner"`
	IsPending types.Bool   `tfsdk:"is_pending"`
	Settings  types.Object `tfsdk:"settings"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an n8n user. This resource allows you to create, read, update, " +
			"and delete users in your n8n instance.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "User identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "User email address. This is required and must be unique.",
				Required:            true,
			},
			"first_name": schema.StringAttribute{
				MarkdownDescription: "User's first name",
				Optional:            true,
			},
			"last_name": schema.StringAttribute{
				MarkdownDescription: "User's last name",
				Optional:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "User role (e.g., 'admin', 'member', 'editor'). If not specified, " +
					"defaults to the instance default role.",
				Optional: true,
				Computed: true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "User password. This is sensitive data and will not be stored in the state after creation.",
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"theme": schema.StringAttribute{
						MarkdownDescription: "User's preferred theme (e.g., 'light', 'dark')",
						Optional:            true,
					},
					"allow_sso_manual_login": schema.BoolAttribute{
						MarkdownDescription: "Whether to allow SSO manual login for this user",
						Optional:            true,
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

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create user request object
	createReq := &client.CreateUserRequest{
		Email:     data.Email.ValueString(),
		FirstName: data.FirstName.ValueString(),
		LastName:  data.LastName.ValueString(),
		Role:      data.Role.ValueString(),
		Password:  data.Password.ValueString(),
	}

	// Create user via API
	createdUser, err := r.client.CreateUser(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user, got error: %s", err))
		return
	}

	// Fetch complete user data after creation (creation response may not include all fields)
	completeUser, err := r.client.GetUser(createdUser.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created user, got error: %s", err))
		return
	}

	// Update model with complete user data
	r.updateModelFromUser(&data, completeUser)

	// Keep password in state (it's marked as sensitive, so it's secure)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get user from API
	user, err := r.client.GetUser(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}

	// Preserve the existing password from state (API doesn't return passwords)
	existingPassword := data.Password

	// Update model with response data
	r.updateModelFromUser(&data, user)

	// Restore the password field
	data.Password = existingPassword

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create user object for update
	user := &client.User{
		Email:     data.Email.ValueString(),
		FirstName: data.FirstName.ValueString(),
		LastName:  data.LastName.ValueString(),
		Role:      data.Role.ValueString(),
	}

	// Handle settings if provided
	if !data.Settings.IsNull() {
		var settings client.UserSettings
		resp.Diagnostics.Append(data.Settings.As(ctx, &settings, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		user.Settings = settings
	}

	// Update user via API
	updatedUser, err := r.client.UpdateUser(data.ID.ValueString(), user)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user, got error: %s", err))
		return
	}

	// Update model with response data
	r.updateModelFromUser(&data, updatedUser)

	// Clear the password from state for security (it's not returned by the API)
	data.Password = types.StringNull()

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete user via API
	err := r.client.DeleteUser(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest,
	resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to update model from API response
func (r *UserResource) updateModelFromUser(model *UserResourceModel, user *client.User) {
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

	// Handle settings (always set to ensure known value)
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

	if user.CreatedAt != nil {
		model.CreatedAt = types.StringValue(user.CreatedAt.Format("2006-01-02T15:04:05Z"))
	} else {
		model.CreatedAt = types.StringNull()
	}

	if user.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(user.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	} else {
		model.UpdatedAt = types.StringNull()
	}
}
