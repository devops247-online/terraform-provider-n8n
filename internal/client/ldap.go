package client

import (
	"fmt"
)

// LDAPConfig represents LDAP configuration (Enterprise feature)
type LDAPConfig struct {
	ServerURL              string `json:"serverUrl"`
	BindDN                 string `json:"bindDn"`
	BindPassword           string `json:"bindPassword"`
	SearchBase             string `json:"searchBase,omitempty"`
	SearchFilter           string `json:"searchFilter,omitempty"`
	UserIDAttribute        string `json:"userIdAttribute,omitempty"`
	UserEmailAttribute     string `json:"userEmailAttribute,omitempty"`
	UserFirstNameAttribute string `json:"userFirstNameAttribute,omitempty"`
	UserLastNameAttribute  string `json:"userLastNameAttribute,omitempty"`
	GroupSearchBase        string `json:"groupSearchBase,omitempty"`
	GroupSearchFilter      string `json:"groupSearchFilter,omitempty"`
	TLSEnabled             bool   `json:"tlsEnabled,omitempty"`
	CACertificate          string `json:"caCertificate,omitempty"`
}

// LDAPTestResult represents the result of testing LDAP connection
type LDAPTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// GetLDAPConfig retrieves the current LDAP configuration
func (c *Client) GetLDAPConfig() (*LDAPConfig, error) {
	var config LDAPConfig
	err := c.Get("ldap/config", &config)
	if err != nil {
		return nil, fmt.Errorf("failed to get LDAP config: %w", err)
	}

	return &config, nil
}

// UpdateLDAPConfig updates the LDAP configuration
func (c *Client) UpdateLDAPConfig(config *LDAPConfig) (*LDAPConfig, error) {
	if config == nil {
		return nil, fmt.Errorf("LDAP config is required")
	}

	if config.ServerURL == "" {
		return nil, fmt.Errorf("LDAP server URL is required")
	}

	if config.BindDN == "" {
		return nil, fmt.Errorf("LDAP bind DN is required")
	}

	if config.BindPassword == "" {
		return nil, fmt.Errorf("LDAP bind password is required")
	}

	var result LDAPConfig
	err := c.Put("ldap/config", config, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update LDAP config: %w", err)
	}

	return &result, nil
}

// TestLDAPConnection tests the LDAP connection with the current configuration
func (c *Client) TestLDAPConnection() (*LDAPTestResult, error) {
	var result LDAPTestResult
	err := c.Post("ldap/test", nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to test LDAP connection: %w", err)
	}

	return &result, nil
}

// TestLDAPConnectionWithConfig tests the LDAP connection with a specific configuration
func (c *Client) TestLDAPConnectionWithConfig(config *LDAPConfig) (*LDAPTestResult, error) {
	if config == nil {
		return nil, fmt.Errorf("LDAP config is required")
	}

	var result LDAPTestResult
	err := c.Post("ldap/test", config, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to test LDAP connection: %w", err)
	}

	return &result, nil
}
