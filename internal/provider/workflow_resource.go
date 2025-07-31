package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/devops247-online/terraform-provider-n8n/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &WorkflowResource{}
var _ resource.ResourceWithImportState = &WorkflowResource{}

func NewWorkflowResource() resource.Resource {
	return &WorkflowResource{}
}

// WorkflowResource defines the resource implementation.
type WorkflowResource struct {
	client *client.Client
}

// WorkflowResourceModel describes the resource data model.
type WorkflowResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Active      types.Bool   `tfsdk:"active"`
	Nodes       types.String `tfsdk:"nodes"`
	Connections types.String `tfsdk:"connections"`
	Settings    types.String `tfsdk:"settings"`
	StaticData  types.String `tfsdk:"static_data"`
	PinnedData  types.String `tfsdk:"pinned_data"`
	Tags        types.List   `tfsdk:"tags"`
	VersionID   types.String `tfsdk:"version_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *WorkflowResource) Metadata(ctx context.Context, req resource.MetadataRequest,
	resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

func (r *WorkflowResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an n8n workflow. Workflows are the core automation units in " +
			"n8n that define a series of nodes and their connections.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Workflow identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the workflow",
				Required:            true,
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the workflow is active and can be triggered",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"nodes": schema.StringAttribute{
				MarkdownDescription: "JSON string containing the workflow nodes configuration",
				Optional:            true,
				Computed:            true,
			},
			"connections": schema.StringAttribute{
				MarkdownDescription: "JSON string containing the workflow connections between nodes",
				Optional:            true,
				Computed:            true,
			},
			"settings": schema.StringAttribute{
				MarkdownDescription: "JSON string containing workflow settings",
				Optional:            true,
				Computed:            true,
			},
			"static_data": schema.StringAttribute{
				MarkdownDescription: "JSON string containing static data for the workflow",
				Optional:            true,
				Computed:            true,
			},
			"pinned_data": schema.StringAttribute{
				MarkdownDescription: "JSON string containing pinned data for testing purposes",
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "List of tags associated with the workflow",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"version_id": schema.StringAttribute{
				MarkdownDescription: "Version identifier of the workflow",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the workflow was created",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the workflow was last updated",
				Computed:            true,
			},
		},
	}
}

func (r *WorkflowResource) Configure(ctx context.Context, req resource.ConfigureRequest,
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

func (r *WorkflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkflowResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create workflow object
	workflow := &client.Workflow{
		Name:   data.Name.ValueString(),
		Active: data.Active.ValueBool(),
	}

	// Parse and validate JSON fields if provided
	if !data.Nodes.IsNull() && data.Nodes.ValueString() != "" {
		if err := r.validateWorkflowJSON(data.Nodes.ValueString(), "nodes"); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("nodes"),
				"Invalid Nodes JSON",
				err.Error(),
			)
			return
		}
		var nodes map[string]interface{}
		if err := json.Unmarshal([]byte(data.Nodes.ValueString()), &nodes); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("nodes"),
				"Invalid JSON",
				fmt.Sprintf("Unable to parse nodes JSON: %s", err),
			)
			return
		}
		// Convert nodes from object format to array format for API
		nodesArray := r.convertNodesToArray(nodes)
		workflow.Nodes = nodesArray
	}

	// Connections field is required by n8n API, default to empty object if not provided
	if !data.Connections.IsNull() && data.Connections.ValueString() != "" {
		if err := r.validateWorkflowJSON(data.Connections.ValueString(), "connections"); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("connections"),
				"Invalid Connections JSON",
				err.Error(),
			)
			return
		}
		var connections map[string]interface{}
		if err := json.Unmarshal([]byte(data.Connections.ValueString()), &connections); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("connections"),
				"Invalid JSON",
				fmt.Sprintf("Unable to parse connections JSON: %s", err),
			)
			return
		}
		workflow.Connections = connections
	} else {
		// Set empty connections object if not provided (required by n8n API)
		workflow.Connections = make(map[string]interface{})
	}

	// Settings field is required by n8n API, default to basic settings if not provided
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
		workflow.Settings = settings
	} else {
		// Set basic settings if not provided (required by n8n API)
		workflow.Settings = map[string]interface{}{
			"executionOrder": "v1",
		}
	}

	if !data.StaticData.IsNull() && data.StaticData.ValueString() != "" {
		var staticData map[string]interface{}
		if err := json.Unmarshal([]byte(data.StaticData.ValueString()), &staticData); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("static_data"),
				"Invalid JSON",
				fmt.Sprintf("Unable to parse static_data JSON: %s", err),
			)
			return
		}
		workflow.StaticData = staticData
	}

	if !data.PinnedData.IsNull() && data.PinnedData.ValueString() != "" {
		var pinnedData map[string]interface{}
		if err := json.Unmarshal([]byte(data.PinnedData.ValueString()), &pinnedData); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("pinned_data"),
				"Invalid JSON",
				fmt.Sprintf("Unable to parse pinned_data JSON: %s", err),
			)
			return
		}
		workflow.PinnedData = pinnedData
	}

	// Tags are read-only during creation, will be set via update if needed

	// Create workflow via API
	createdWorkflow, err := r.client.CreateWorkflow(workflow)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create workflow, got error: %s", err))
		return
	}

	// TODO: Tags are read-only in n8n API, need to investigate proper tag management approach

	// Update model with response data
	r.updateModelFromWorkflow(&data, createdWorkflow)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkflowResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get workflow from API
	workflow, err := r.client.GetWorkflow(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read workflow, got error: %s", err))
		return
	}

	// Update model with response data
	r.updateModelFromWorkflow(&data, workflow)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WorkflowResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create workflow object for update
	workflow := &client.Workflow{
		Name:   data.Name.ValueString(),
		Active: data.Active.ValueBool(),
	}

	// Parse and validate JSON fields if provided (similar to Create method)
	if !data.Nodes.IsNull() && data.Nodes.ValueString() != "" {
		if err := r.validateWorkflowJSON(data.Nodes.ValueString(), "nodes"); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("nodes"),
				"Invalid Nodes JSON",
				err.Error(),
			)
			return
		}
		var nodes map[string]interface{}
		if err := json.Unmarshal([]byte(data.Nodes.ValueString()), &nodes); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("nodes"),
				"Invalid JSON",
				fmt.Sprintf("Unable to parse nodes JSON: %s", err),
			)
			return
		}
		// Convert nodes from object format to array format for API
		nodesArray := r.convertNodesToArray(nodes)
		workflow.Nodes = nodesArray
	}

	// Connections field is required by n8n API, default to empty object if not provided
	if !data.Connections.IsNull() && data.Connections.ValueString() != "" {
		if err := r.validateWorkflowJSON(data.Connections.ValueString(), "connections"); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("connections"),
				"Invalid Connections JSON",
				err.Error(),
			)
			return
		}
		var connections map[string]interface{}
		if err := json.Unmarshal([]byte(data.Connections.ValueString()), &connections); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("connections"),
				"Invalid JSON",
				fmt.Sprintf("Unable to parse connections JSON: %s", err),
			)
			return
		}
		workflow.Connections = connections
	} else {
		// Set empty connections object if not provided (required by n8n API)
		workflow.Connections = make(map[string]interface{})
	}

	// Settings field is required by n8n API, default to basic settings if not provided
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
		workflow.Settings = settings
	} else {
		// Set basic settings if not provided (required by n8n API)
		workflow.Settings = map[string]interface{}{
			"executionOrder": "v1",
		}
	}

	if !data.StaticData.IsNull() && data.StaticData.ValueString() != "" {
		var staticData map[string]interface{}
		if err := json.Unmarshal([]byte(data.StaticData.ValueString()), &staticData); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("static_data"),
				"Invalid JSON",
				fmt.Sprintf("Unable to parse static_data JSON: %s", err),
			)
			return
		}
		workflow.StaticData = staticData
	}

	if !data.PinnedData.IsNull() && data.PinnedData.ValueString() != "" {
		var pinnedData map[string]interface{}
		if err := json.Unmarshal([]byte(data.PinnedData.ValueString()), &pinnedData); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("pinned_data"),
				"Invalid JSON",
				fmt.Sprintf("Unable to parse pinned_data JSON: %s", err),
			)
			return
		}
		workflow.PinnedData = pinnedData
	}

	// Handle tags
	if !data.Tags.IsNull() {
		var tags []string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		workflow.Tags = tags
	}

	// Update workflow via API
	updatedWorkflow, err := r.client.UpdateWorkflow(data.ID.ValueString(), workflow)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update workflow, got error: %s", err))
		return
	}

	// Update model with response data
	r.updateModelFromWorkflow(&data, updatedWorkflow)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkflowResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete workflow via API
	err := r.client.DeleteWorkflow(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete workflow, got error: %s", err))
		return
	}
}

func (r *WorkflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest,
	resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// validateWorkflowJSON validates the JSON structure of workflow fields
func (r *WorkflowResource) validateWorkflowJSON(jsonStr string, fieldName string) error {
	if jsonStr == "" {
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return fmt.Errorf("invalid JSON in %s: %w", fieldName, err)
	}

	// Additional validation for specific fields
	switch fieldName {
	case "nodes":
		// Validate nodes structure - should be a map where each key represents a node
		for nodeKey, nodeValue := range result {
			if nodeMap, ok := nodeValue.(map[string]interface{}); ok {
				// Check for required node properties
				if _, hasType := nodeMap["type"]; !hasType {
					return fmt.Errorf("node %s is missing required 'type' field", nodeKey)
				}
			} else {
				return fmt.Errorf("node %s must be an object", nodeKey)
			}
		}
	case "connections":
		// Validate connections structure - should be a map of arrays
		for sourceNode, connections := range result {
			if connArray, ok := connections.(map[string]interface{}); ok {
				for outputType, outputConnections := range connArray {
					if connectionsList, ok := outputConnections.([]interface{}); ok {
						for i, conn := range connectionsList {
							if connMap, ok := conn.(map[string]interface{}); ok {
								if _, hasNode := connMap["node"]; !hasNode {
									return fmt.Errorf("connection %d from %s.%s is missing required 'node' field", i, sourceNode, outputType)
								}
								if _, hasType := connMap["type"]; !hasType {
									return fmt.Errorf("connection %d from %s.%s is missing required 'type' field", i, sourceNode, outputType)
								}
								if _, hasIndex := connMap["index"]; !hasIndex {
									return fmt.Errorf("connection %d from %s.%s is missing required 'index' field", i, sourceNode, outputType)
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// Helper function to update model from API response
func (r *WorkflowResource) updateModelFromWorkflow(model *WorkflowResourceModel, workflow *client.Workflow) {
	model.ID = types.StringValue(workflow.ID)
	model.Name = types.StringValue(workflow.Name)
	model.Active = types.BoolValue(workflow.Active)

	// Convert JSON fields to strings
	if workflow.Nodes != nil {
		// Convert nodes from API array format to Terraform object format
		nodesObject := r.convertNodesFromArray(workflow.Nodes)
		if nodesJSON, err := json.Marshal(nodesObject); err == nil {
			model.Nodes = types.StringValue(string(nodesJSON))
		}
	}

	if workflow.Connections != nil {
		if connectionsJSON, err := json.Marshal(workflow.Connections); err == nil {
			model.Connections = types.StringValue(string(connectionsJSON))
		}
	}

	if workflow.Settings != nil {
		if settingsJSON, err := json.Marshal(workflow.Settings); err == nil {
			model.Settings = types.StringValue(string(settingsJSON))
		}
	}

	if workflow.StaticData != nil {
		if staticDataJSON, err := json.Marshal(workflow.StaticData); err == nil {
			model.StaticData = types.StringValue(string(staticDataJSON))
		}
	} else {
		model.StaticData = types.StringNull()
	}

	if workflow.PinnedData != nil {
		if pinnedDataJSON, err := json.Marshal(workflow.PinnedData); err == nil {
			model.PinnedData = types.StringValue(string(pinnedDataJSON))
		}
	} else {
		model.PinnedData = types.StringNull()
	}

	// Handle tags
	if workflow.Tags != nil {
		tagValues := make([]attr.Value, len(workflow.Tags))
		for i, tag := range workflow.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		model.Tags = types.ListValueMust(types.StringType, tagValues)
	}

	if workflow.VersionID != "" {
		model.VersionID = types.StringValue(workflow.VersionID)
	}

	if workflow.CreatedAt != nil {
		model.CreatedAt = types.StringValue(workflow.CreatedAt.Format("2006-01-02T15:04:05Z"))
	}

	if workflow.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(workflow.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	}
}

// convertNodesToArray converts nodes from Terraform's object format to n8n API's array format
func (r *WorkflowResource) convertNodesToArray(nodes map[string]interface{}) []interface{} {
	var nodesArray []interface{}

	for nodeId, nodeData := range nodes {
		if nodeMap, ok := nodeData.(map[string]interface{}); ok {
			// Add the node ID to the node data
			nodeMap["id"] = nodeId
			nodesArray = append(nodesArray, nodeMap)
		}
	}

	return nodesArray
}

// convertNodesFromArray converts nodes from n8n API's array format to Terraform's object format
func (r *WorkflowResource) convertNodesFromArray(nodesArray []interface{}) map[string]interface{} {
	nodesObject := make(map[string]interface{})

	for _, nodeData := range nodesArray {
		if nodeMap, ok := nodeData.(map[string]interface{}); ok {
			if nodeId, exists := nodeMap["id"]; exists {
				if nodeIdStr, ok := nodeId.(string); ok {
					// Remove the id field from the node data since it becomes the key
					nodeCopy := make(map[string]interface{})
					for k, v := range nodeMap {
						if k != "id" {
							nodeCopy[k] = v
						}
					}
					nodesObject[nodeIdStr] = nodeCopy
				}
			}
		}
	}

	return nodesObject
}
