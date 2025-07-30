package client

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// User represents an n8n user
type User struct {
	ID          string       `json:"id,omitempty"`
	Email       string       `json:"email"`
	FirstName   string       `json:"firstName,omitempty"`
	LastName    string       `json:"lastName,omitempty"`
	Role        string       `json:"role,omitempty"`
	IsOwner     bool         `json:"isOwner,omitempty"`
	IsPending   bool         `json:"isPending,omitempty"`
	SignupToken string       `json:"signupToken,omitempty"`
	Settings    UserSettings `json:"settings,omitempty"`
	CreatedAt   *time.Time   `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time   `json:"updatedAt,omitempty"`
}

// UserSettings represents user-specific settings
type UserSettings struct {
	Theme               string `json:"theme,omitempty"`
	AllowSSOManualLogin bool   `json:"allowSSOManualLogin,omitempty"`
}

// UserListOptions represents options for listing users
type UserListOptions struct {
	Role   string
	Limit  int
	Offset int
}

// UserListResponse represents the response from listing users
type UserListResponse struct {
	Data       []User `json:"data"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Role      string `json:"role,omitempty"`
	Password  string `json:"password,omitempty"`
}

// GetUsers retrieves a list of users
func (c *Client) GetUsers(options *UserListOptions) (*UserListResponse, error) {
	u, err := url.Parse("users")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if options != nil {
		params := url.Values{}

		if options.Role != "" {
			params.Set("role", options.Role)
		}

		if options.Limit > 0 {
			params.Set("limit", strconv.Itoa(options.Limit))
		}

		if options.Offset > 0 {
			params.Set("offset", strconv.Itoa(options.Offset))
		}

		u.RawQuery = params.Encode()
	}

	var result UserListResponse
	err = c.Get(u.String(), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return &result, nil
}

// GetUser retrieves a specific user by ID
func (c *Client) GetUser(id string) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	path := fmt.Sprintf("users/%s", id)

	var user User
	err := c.Get(path, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", id, err)
	}

	return &user, nil
}

// CreateUser creates a new user
func (c *Client) CreateUser(userReq *CreateUserRequest) (*User, error) {
	if userReq == nil {
		return nil, fmt.Errorf("user request is required")
	}

	if userReq.Email == "" {
		return nil, fmt.Errorf("user email is required")
	}

	// n8n API expects an array of users, so wrap single user in array
	userArray := []*CreateUserRequest{userReq}
	
	// n8n returns array of {user: User, error: string} objects
	type CreateUserResponse struct {
		User  User   `json:"user"`
		Error string `json:"error"`
	}
	
	var resultArray []CreateUserResponse
	err := c.Post("users", userArray, &resultArray)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	if len(resultArray) == 0 {
		return nil, fmt.Errorf("no user returned from API")
	}
	
	if resultArray[0].Error != "" {
		return nil, fmt.Errorf("user creation failed: %s", resultArray[0].Error)
	}

	return &resultArray[0].User, nil
}

// UpdateUser updates an existing user
func (c *Client) UpdateUser(id string, user *User) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if user == nil {
		return nil, fmt.Errorf("user is required")
	}

	path := fmt.Sprintf("users/%s", id)

	var result User
	err := c.Put(path, user, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update user %s: %w", id, err)
	}

	return &result, nil
}

// DeleteUser deletes a user
func (c *Client) DeleteUser(id string) error {
	if id == "" {
		return fmt.Errorf("user ID is required")
	}

	path := fmt.Sprintf("users/%s", id)

	err := c.Delete(path)
	if err != nil {
		return fmt.Errorf("failed to delete user %s: %w", id, err)
	}

	return nil
}
