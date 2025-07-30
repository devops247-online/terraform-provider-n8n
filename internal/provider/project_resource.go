package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/devops247-online/terraform-provider-n8n/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

// ProjectResource defines the resource implementation.
type ProjectResource struct {
	client *client.Client
}

// ProjectResourceModel describes the resource data model.
type ProjectResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Settings    types.String `tfsdk:"settings"`
	Icon        types.String `tfsdk:"icon"`
	Color       types.String `tfsdk:"color"`
	OwnerID     types.String `tfsdk:"owner_id"`
	MemberCount types.Int64  `tfsdk:"member_count"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *ProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an n8n project. Projects provide workspace isolation and team collaboration features in n8n Enterprise.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Project identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the project",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the project",
				Optional:            true,
			},
			"settings": schema.StringAttribute{
				MarkdownDescription: "JSON string containing project-specific settings",
				Optional:            true,
				Computed:            true,
			},
			"icon": schema.StringAttribute{
				MarkdownDescription: "Project icon identifier",
				Optional:            true,
			},
			"color": schema.StringAttribute{
				MarkdownDescription: "Project color scheme",
				Optional:            true,
			},
			"owner_id": schema.StringAttribute{
				MarkdownDescription: "Project owner user ID",
				Computed:            true,
			},
			"member_count": schema.Int64Attribute{
				MarkdownDescription: "Number of project members",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the project was created",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the project was last updated",
				Computed:            true,
			},
		},
	}
}

func (r *ProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create project object
	project := &client.Project{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Icon:        data.Icon.ValueString(),
		Color:       data.Color.ValueString(),
	}

	// Parse and validate settings JSON if provided
	if !data.Settings.IsNull() && data.Settings.ValueString() != "" {
		var settings map[string]interface{}
		if err := json.Unmarshal([]byte(data.Settings.ValueString()), &settings); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("settings"),
				"Invalid JSON",
				fmt.Sprintf("Unable to parse settings JSON: %s", err),
			)
			return
		}
		project.Settings = settings
	}

	// Create project via API
	createdProject, err := r.client.CreateProject(project)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create project, got error: %s", err))
		return
	}

	// Update model with response data
	r.updateModelFromProject(&data, createdProject)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get project from API
	project, err := r.client.GetProject(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read project, got error: %s", err))
		return
	}

	// Update model with response data
	r.updateModelFromProject(&data, project)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create project object for update
	project := &client.Project{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Icon:        data.Icon.ValueString(),
		Color:       data.Color.ValueString(),
	}

	// Parse and validate settings JSON if provided
	if !data.Settings.IsNull() && data.Settings.ValueString() != "" {
		var settings map[string]interface{}
		if err := json.Unmarshal([]byte(data.Settings.ValueString()), &settings); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("settings"),
				"Invalid JSON",
				fmt.Sprintf("Unable to parse settings JSON: %s", err),
			)
			return
		}
		project.Settings = settings
	}

	// Update project via API
	updatedProject, err := r.client.UpdateProject(data.ID.ValueString(), project)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update project, got error: %s", err))
		return
	}

	// Update model with response data
	r.updateModelFromProject(&data, updatedProject)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete project via API
	err := r.client.DeleteProject(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete project, got error: %s", err))
		return
	}
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to update model from API response
func (r *ProjectResource) updateModelFromProject(model *ProjectResourceModel, project *client.Project) {
	model.ID = types.StringValue(project.ID)
	model.Name = types.StringValue(project.Name)
	model.Description = types.StringValue(project.Description)
	model.Icon = types.StringValue(project.Icon)
	model.Color = types.StringValue(project.Color)
	model.OwnerID = types.StringValue(project.OwnerID)
	model.MemberCount = types.Int64Value(int64(project.MemberCount))

	// Convert settings to JSON string
	if project.Settings != nil {
		if settingsJSON, err := json.Marshal(project.Settings); err == nil {
			model.Settings = types.StringValue(string(settingsJSON))
		}
	}

	if project.CreatedAt != nil {
		model.CreatedAt = types.StringValue(project.CreatedAt.Format("2006-01-02T15:04:05Z"))
	}

	if project.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(project.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	}
}
