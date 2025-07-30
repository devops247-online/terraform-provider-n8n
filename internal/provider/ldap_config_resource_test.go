package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLDAPConfigResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLDAPConfigResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "server_url", "ldap://ldap.example.com:389"),
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "bind_dn", "cn=admin,dc=example,dc=com"),
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "search_base", "ou=users,dc=example,dc=com"),
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "search_filter", "(uid={{username}})"),
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "user_id_attribute", "uid"),
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "user_email_attribute", "mail"),
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "tls_enabled", "false"),
					resource.TestCheckResourceAttrSet("n8n_ldap_config.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "n8n_ldap_config.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bind_password"}, // Password is sensitive and not returned
			},
			// Update and Read testing
			{
				Config: testAccLDAPConfigResourceConfigUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "server_url", "ldaps://ldap.example.com:636"),
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "search_base", "ou=people,dc=example,dc=com"),
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "tls_enabled", "true"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccLDAPConfigResource_WithTLS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with TLS
			{
				Config: testAccLDAPConfigResourceConfigWithTLS(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "server_url", "ldaps://secure-ldap.example.com:636"),
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "tls_enabled", "true"),
					resource.TestCheckResourceAttrSet("n8n_ldap_config.test", "ca_certificate"),
				),
			},
		},
	})
}

func TestAccLDAPConfigResource_WithGroups(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with group configuration
			{
				Config: testAccLDAPConfigResourceConfigWithGroups(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "group_search_base", "ou=groups,dc=example,dc=com"),
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "group_search_filter", "(member={{userDN}})"),
				),
			},
		},
	})
}

func TestAccLDAPConfigResource_MinimalConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with minimal required config
			{
				Config: testAccLDAPConfigResourceConfigMinimal(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "server_url", "ldap://minimal.example.com:389"),
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "bind_dn", "cn=bind,dc=example,dc=com"),
					// Check computed defaults
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "user_id_attribute", "uid"),
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "user_email_attribute", "mail"),
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "search_filter", "(uid={{username}})"),
					resource.TestCheckResourceAttr("n8n_ldap_config.test", "tls_enabled", "false"),
				),
			},
		},
	})
}

func testAccLDAPConfigResourceConfig() string {
	return `
resource "n8n_ldap_config" "test" {
  server_url                = "ldap://ldap.example.com:389"
  bind_dn                   = "cn=admin,dc=example,dc=com"
  bind_password             = "secret123"
  search_base               = "ou=users,dc=example,dc=com"
  search_filter             = "(uid={{username}})"
  user_id_attribute         = "uid"
  user_email_attribute      = "mail"
  user_first_name_attribute = "givenName"
  user_last_name_attribute  = "sn"
  tls_enabled               = false
}
`
}

func testAccLDAPConfigResourceConfigUpdated() string {
	return `
resource "n8n_ldap_config" "test" {
  server_url                = "ldaps://ldap.example.com:636"
  bind_dn                   = "cn=admin,dc=example,dc=com"
  bind_password             = "newsecret456"
  search_base               = "ou=people,dc=example,dc=com"
  search_filter             = "(cn={{username}})"
  user_id_attribute         = "cn"
  user_email_attribute      = "email"
  user_first_name_attribute = "givenName"
  user_last_name_attribute  = "surname"
  tls_enabled               = true
}
`
}

func testAccLDAPConfigResourceConfigWithTLS() string {
	return `
resource "n8n_ldap_config" "test" {
  server_url    = "ldaps://secure-ldap.example.com:636"
  bind_dn       = "cn=secure-bind,dc=example,dc=com"
  bind_password = "securepass789"
  search_base   = "ou=users,dc=example,dc=com"
  tls_enabled   = true
  ca_certificate = <<-EOT
-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKoK/heBjcOuMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTcwNTEwMTk0MDA2WhcNMTgwNTEwMTk0MDA2WjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAuBdKTOA01h5X2sJK22vqXGE9YzfU+L/7L0KOwBCqJYvr3nRPQ8u7JCnZ
example-ca-certificate-content
-----END CERTIFICATE-----
EOT
}
`
}

func testAccLDAPConfigResourceConfigWithGroups() string {
	return `
resource "n8n_ldap_config" "test" {
  server_url            = "ldap://ldap.example.com:389"
  bind_dn               = "cn=admin,dc=example,dc=com"
  bind_password         = "secret123"
  search_base           = "ou=users,dc=example,dc=com"
  group_search_base     = "ou=groups,dc=example,dc=com"
  group_search_filter   = "(member={{userDN}})"
}
`
}

func testAccLDAPConfigResourceConfigMinimal() string {
	return `
resource "n8n_ldap_config" "test" {
  server_url    = "ldap://minimal.example.com:389"
  bind_dn       = "cn=bind,dc=example,dc=com"
  bind_password = "minimalpass"
}
`
}
