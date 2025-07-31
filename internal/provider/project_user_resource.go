package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/devops247-online/terraform-provider-n8n/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ProjectUserResource{}
var _ resource.ResourceWithImportState = &ProjectUserResource{}

func NewProjectUserResource() resource.Resource {
	return &ProjectUserResource{}
}

// ProjectUserResource defines the resource implementation.
type ProjectUserResource struct {
	client *client.Client
}

// ProjectUserResourceModel describes the resource data model.
type ProjectUserResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	UserID    types.String `tfsdk:"user_id"`
	Role      types.String `tfsdk:"role"`
	AddedAt   types.String `tfsdk:"added_at"`
}

func (r *ProjectUserResource) Metadata(ctx context.Context, req resource.MetadataRequest,
	resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_user"
}

func (r *ProjectUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages user membership in an n8n project. This resource allows you to " +
			"assign users to projects with specific roles.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Project user assignment identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "The ID or email of the user to add to the project",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role of the user in the project (admin, editor, viewer)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("viewer"),
			},
			"added_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the user was added to the project",
				Computed:            true,
			},
		},
	}
}

func (r *ProjectUserResource) Configure(ctx context.Context, req resource.ConfigureRequest,
	resp *resource.ConfigureResponse) {
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

func (r *ProjectUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectUserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create project user object
	projectUser := &client.ProjectUser{
		ProjectID: data.ProjectID.ValueString(),
		UserID:    data.UserID.ValueString(),
		Role:      data.Role.ValueString(),
	}

	// Add user to project via API
	createdProjectUser, err := r.client.AddUserToProject(projectUser)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add user to project, got error: %s", err))
		return
	}

	// Update model with response data
	r.updateModelFromProjectUser(&data, createdProjectUser)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectUserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get project users from API
	projectUsers, err := r.client.GetProjectUsers(data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read project users, got error: %s", err))
		return
	}

	// Find the specific user
	var foundUser *client.ProjectUser
	for _, user := range projectUsers {
		if user.UserID == data.UserID.ValueString() {
			foundUser = &user
			break
		}
	}

	if foundUser == nil {
		resp.Diagnostics.AddError("Not Found",
			fmt.Sprintf("User %s not found in project %s", data.UserID.ValueString(), data.ProjectID.ValueString()))
		return
	}

	// Update model with response data
	r.updateModelFromProjectUser(&data, foundUser)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectUserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create project user object for update
	projectUser := &client.ProjectUser{
		ProjectID: data.ProjectID.ValueString(),
		UserID:    data.UserID.ValueString(),
		Role:      data.Role.ValueString(),
	}

	// Update project user via API
	updatedProjectUser, err := r.client.UpdateProjectUser(data.ProjectID.ValueString(),
		data.UserID.ValueString(), projectUser)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update project user, got error: %s", err))
		return
	}

	// Update model with response data
	r.updateModelFromProjectUser(&data, updatedProjectUser)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectUserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Remove user from project via API
	err := r.client.RemoveUserFromProject(data.ProjectID.ValueString(), data.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove user from project, got error: %s", err))
		return
	}
}

func (r *ProjectUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest,
	resp *resource.ImportStateResponse) {
	// Import state should be in the format "project_id:user_id"
	// We'll parse this in the ID field and then set the individual fields
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to update model from API response
func (r *ProjectUserResource) updateModelFromProjectUser(model *ProjectUserResourceModel,
	projectUser *client.ProjectUser) {
	model.ID = types.StringValue(fmt.Sprintf("%s:%s", projectUser.ProjectID, projectUser.UserID))
	model.ProjectID = types.StringValue(projectUser.ProjectID)
	model.UserID = types.StringValue(projectUser.UserID)
	model.Role = types.StringValue(projectUser.Role)

	if projectUser.AddedAt != nil {
		model.AddedAt = types.StringValue(projectUser.AddedAt.Format("2006-01-02T15:04:05Z"))
	}
}
