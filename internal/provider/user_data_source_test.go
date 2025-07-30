package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing by ID
			{
				Config: testAccUserDataSourceConfig("n8n_user.test.id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.n8n_user.test", "email", "datasource@example.com"),
					resource.TestCheckResourceAttr("data.n8n_user.test", "first_name", "DataSource"),
					resource.TestCheckResourceAttr("data.n8n_user.test", "last_name", "Test"),
					resource.TestCheckResourceAttr("data.n8n_user.test", "role", "member"),
					resource.TestCheckResourceAttrSet("data.n8n_user.test", "id"),
					resource.TestCheckResourceAttrSet("data.n8n_user.test", "created_at"),
				),
			},
		},
	})
}

func TestAccUserDataSourceByEmail(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing by email
			{
				Config: testAccUserDataSourceByEmailConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.n8n_user.test", "email", "datasource-email@example.com"),
					resource.TestCheckResourceAttr("data.n8n_user.test", "first_name", "Email"),
					resource.TestCheckResourceAttr("data.n8n_user.test", "last_name", "Lookup"),
					resource.TestCheckResourceAttr("data.n8n_user.test", "role", "admin"),
					resource.TestCheckResourceAttrSet("data.n8n_user.test", "id"),
				),
			},
		},
	})
}

func testAccUserDataSourceConfig(idReference string) string {
	return fmt.Sprintf(`
resource "n8n_user" "test" {
  email      = "datasource@example.com"
  first_name = "DataSource"
  last_name  = "Test"
  role       = "member"
}

data "n8n_user" "test" {
  id = %s
}
`, idReference)
}

func testAccUserDataSourceByEmailConfig() string {
	return `
resource "n8n_user" "test" {
  email      = "datasource-email@example.com"
  first_name = "Email"
  last_name  = "Lookup"
  role       = "admin"
}

data "n8n_user" "test" {
  email = n8n_user.test.email
}
`
}
