package client

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Project represents an n8n project (Enterprise feature)
type Project struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	Icon        string                 `json:"icon,omitempty"`
	Color       string                 `json:"color,omitempty"`
	OwnerID     string                 `json:"ownerId,omitempty"`
	MemberCount int                    `json:"memberCount,omitempty"`
	CreatedAt   *time.Time             `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time             `json:"updatedAt,omitempty"`
}

// ProjectUser represents a user's membership in a project
type ProjectUser struct {
	ID        string     `json:"id,omitempty"`
	ProjectID string     `json:"projectId"`
	UserID    string     `json:"userId"`
	Role      string     `json:"role,omitempty"`
	AddedAt   *time.Time `json:"addedAt,omitempty"`
}

// ProjectListOptions represents options for listing projects
type ProjectListOptions struct {
	Limit  int
	Offset int
}

// ProjectListResponse represents the response from listing projects
type ProjectListResponse struct {
	Data       []Project `json:"data"`
	NextCursor string    `json:"nextCursor,omitempty"`
}

// GetProjects retrieves a list of projects
func (c *Client) GetProjects(options *ProjectListOptions) (*ProjectListResponse, error) {
	path := "projects"

	if options != nil {
		params := url.Values{}

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

	var result ProjectListResponse
	err := c.Get(path, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return &result, nil
}

// GetProject retrieves a specific project by ID
func (c *Client) GetProject(id string) (*Project, error) {
	if id == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	path := fmt.Sprintf("projects/%s", id)

	var project Project
	err := c.Get(path, &project)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", id, err)
	}

	return &project, nil
}

// CreateProject creates a new project
func (c *Client) CreateProject(project *Project) (*Project, error) {
	if project == nil {
		return nil, fmt.Errorf("project is required")
	}

	if project.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	var result Project
	err := c.Post("projects", project, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &result, nil
}

// UpdateProject updates an existing project
func (c *Client) UpdateProject(id string, project *Project) (*Project, error) {
	if id == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	if project == nil {
		return nil, fmt.Errorf("project is required")
	}

	path := fmt.Sprintf("projects/%s", id)

	var result Project
	err := c.Put(path, project, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update project %s: %w", id, err)
	}

	return &result, nil
}

// DeleteProject deletes a project
func (c *Client) DeleteProject(id string) error {
	if id == "" {
		return fmt.Errorf("project ID is required")
	}

	path := fmt.Sprintf("projects/%s", id)

	err := c.Delete(path)
	if err != nil {
		return fmt.Errorf("failed to delete project %s: %w", id, err)
	}

	return nil
}

// GetProjectUsers retrieves users for a specific project
func (c *Client) GetProjectUsers(projectID string) ([]ProjectUser, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	path := fmt.Sprintf("projects/%s/users", projectID)

	var result struct {
		Data []ProjectUser `json:"data"`
	}
	err := c.Get(path, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get project users for project %s: %w", projectID, err)
	}

	return result.Data, nil
}

// AddUserToProject adds a user to a project
func (c *Client) AddUserToProject(projectUser *ProjectUser) (*ProjectUser, error) {
	if projectUser == nil {
		return nil, fmt.Errorf("project user is required")
	}

	if projectUser.ProjectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	if projectUser.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	path := fmt.Sprintf("projects/%s/users", projectUser.ProjectID)

	var result ProjectUser
	err := c.Post(path, projectUser, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to add user to project: %w", err)
	}

	return &result, nil
}

// UpdateProjectUser updates a user's role in a project
func (c *Client) UpdateProjectUser(projectID, userID string, projectUser *ProjectUser) (*ProjectUser, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if projectUser == nil {
		return nil, fmt.Errorf("project user is required")
	}

	path := fmt.Sprintf("projects/%s/users/%s", projectID, userID)

	var result ProjectUser
	err := c.Put(path, projectUser, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update project user: %w", err)
	}

	return &result, nil
}

// RemoveUserFromProject removes a user from a project
func (c *Client) RemoveUserFromProject(projectID, userID string) error {
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	path := fmt.Sprintf("projects/%s/users/%s", projectID, userID)

	err := c.Delete(path)
	if err != nil {
		return fmt.Errorf("failed to remove user from project: %w", err)
	}

	return nil
}
