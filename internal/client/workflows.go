package client

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Workflow represents an n8n workflow
type Workflow struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name"`
	Active      bool                   `json:"active,omitempty"`
	Nodes       []interface{}          `json:"nodes,omitempty"`
	Connections map[string]interface{} `json:"connections"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	StaticData  map[string]interface{} `json:"staticData,omitempty"`
	PinnedData  map[string]interface{} `json:"pinnedData,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	VersionID   string                 `json:"versionId,omitempty"`
	CreatedAt   *time.Time             `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time             `json:"updatedAt,omitempty"`
}

// WorkflowListOptions represents options for listing workflows
type WorkflowListOptions struct {
	Active    *bool
	Tags      []string
	ProjectID string
	Limit     int
	Offset    int
}

// WorkflowListResponse represents the response from listing workflows
type WorkflowListResponse struct {
	Data       []Workflow `json:"data"`
	NextCursor string     `json:"nextCursor,omitempty"`
}

// GetWorkflows retrieves a list of workflows
func (c *Client) GetWorkflows(options *WorkflowListOptions) (*WorkflowListResponse, error) {
	path := "workflows"

	if options != nil {
		params := url.Values{}

		if options.Active != nil {
			params.Set("active", strconv.FormatBool(*options.Active))
		}

		if len(options.Tags) > 0 {
			for _, tag := range options.Tags {
				params.Add("tags", tag)
			}
		}

		if options.ProjectID != "" {
			params.Set("projectId", options.ProjectID)
		}

		if options.Limit > 0 {
			params.Set("limit", strconv.Itoa(options.Limit))
		}

		if options.Offset > 0 {
			params.Set("offset", strconv.Itoa(options.Offset))
		}

		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var result WorkflowListResponse
	err := c.Get(path, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflows: %w", err)
	}

	return &result, nil
}

// GetWorkflow retrieves a specific workflow by ID
func (c *Client) GetWorkflow(id string) (*Workflow, error) {
	if id == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}

	path := fmt.Sprintf("workflows/%s", id)

	var workflow Workflow
	err := c.Get(path, &workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow %s: %w", id, err)
	}

	return &workflow, nil
}

// CreateWorkflow creates a new workflow
func (c *Client) CreateWorkflow(workflow *Workflow) (*Workflow, error) {
	if workflow == nil {
		return nil, fmt.Errorf("workflow is required")
	}

	if workflow.Name == "" {
		return nil, fmt.Errorf("workflow name is required")
	}

	var result Workflow
	err := c.Post("workflows", workflow, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	return &result, nil
}

// UpdateWorkflow updates an existing workflow
func (c *Client) UpdateWorkflow(id string, workflow *Workflow) (*Workflow, error) {
	if id == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}

	if workflow == nil {
		return nil, fmt.Errorf("workflow is required")
	}

	path := fmt.Sprintf("workflows/%s", id)

	var result Workflow
	err := c.Put(path, workflow, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update workflow %s: %w", id, err)
	}

	return &result, nil
}

// DeleteWorkflow deletes a workflow
func (c *Client) DeleteWorkflow(id string) error {
	if id == "" {
		return fmt.Errorf("workflow ID is required")
	}

	path := fmt.Sprintf("workflows/%s", id)

	err := c.Delete(path)
	if err != nil {
		return fmt.Errorf("failed to delete workflow %s: %w", id, err)
	}

	return nil
}

// ActivateWorkflow activates a workflow
func (c *Client) ActivateWorkflow(id string) (*Workflow, error) {
	if id == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}

	path := fmt.Sprintf("workflows/%s/activate", id)

	var result Workflow
	err := c.Post(path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to activate workflow %s: %w", id, err)
	}

	return &result, nil
}

// DeactivateWorkflow deactivates a workflow
func (c *Client) DeactivateWorkflow(id string) (*Workflow, error) {
	if id == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}

	path := fmt.Sprintf("workflows/%s/deactivate", id)

	var result Workflow
	err := c.Post(path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate workflow %s: %w", id, err)
	}

	return &result, nil
}
