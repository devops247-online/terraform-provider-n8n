package client

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Credential represents an n8n credential
type Credential struct {
	ID         string                 `json:"id,omitempty"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Data       map[string]interface{} `json:"data,omitempty"`
	SharedWith []string               `json:"sharedWith,omitempty"`
	ProjectID  string                 `json:"projectId,omitempty"`
	CreatedAt  *time.Time             `json:"createdAt,omitempty"`
	UpdatedAt  *time.Time             `json:"updatedAt,omitempty"`
}

// CredentialListOptions represents options for listing credentials
type CredentialListOptions struct {
	Type      string
	ProjectID string
	Limit     int
	Offset    int
}

// CredentialListResponse represents the response from listing credentials
type CredentialListResponse struct {
	Data       []Credential `json:"data"`
	NextCursor string       `json:"nextCursor,omitempty"`
}

// GetCredentials retrieves a list of credentials
func (c *Client) GetCredentials(options *CredentialListOptions) (*CredentialListResponse, error) {
	u, err := url.Parse("credentials")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if options != nil {
		params := url.Values{}

		if options.Type != "" {
			params.Set("type", options.Type)
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

		u.RawQuery = params.Encode()
	}

	var result CredentialListResponse
	err = c.Get(u.String(), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	return &result, nil
}

// GetCredential retrieves a specific credential by ID
func (c *Client) GetCredential(id string) (*Credential, error) {
	if id == "" {
		return nil, fmt.Errorf("credential ID is required")
	}

	path := fmt.Sprintf("credentials/%s", id)

	var credential Credential
	err := c.Get(path, &credential)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential %s: %w", id, err)
	}

	return &credential, nil
}

// CreateCredential creates a new credential
func (c *Client) CreateCredential(credential *Credential) (*Credential, error) {
	if credential == nil {
		return nil, fmt.Errorf("credential is required")
	}

	if credential.Name == "" {
		return nil, fmt.Errorf("credential name is required")
	}

	if credential.Type == "" {
		return nil, fmt.Errorf("credential type is required")
	}

	var result Credential
	err := c.Post("credentials", credential, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	return &result, nil
}

// UpdateCredential updates an existing credential
func (c *Client) UpdateCredential(id string, credential *Credential) (*Credential, error) {
	if id == "" {
		return nil, fmt.Errorf("credential ID is required")
	}

	if credential == nil {
		return nil, fmt.Errorf("credential is required")
	}

	path := fmt.Sprintf("credentials/%s", id)

	var result Credential
	err := c.Put(path, credential, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update credential %s: %w", id, err)
	}

	return &result, nil
}

// DeleteCredential deletes a credential
func (c *Client) DeleteCredential(id string) error {
	if id == "" {
		return fmt.Errorf("credential ID is required")
	}

	path := fmt.Sprintf("credentials/%s", id)

	err := c.Delete(path)
	if err != nil {
		return fmt.Errorf("failed to delete credential %s: %w", id, err)
	}

	return nil
}
