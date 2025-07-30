package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserResourceConfig("test@example.com", "Test", "User", "member"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_user.test", "email", "test@example.com"),
					resource.TestCheckResourceAttr("n8n_user.test", "first_name", "Test"),
					resource.TestCheckResourceAttr("n8n_user.test", "last_name", "User"),
					resource.TestCheckResourceAttr("n8n_user.test", "role", "member"),
					resource.TestCheckResourceAttrSet("n8n_user.test", "id"),
					resource.TestCheckResourceAttrSet("n8n_user.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "n8n_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			// Update and Read testing
			{
				Config: testAccUserResourceConfig("test@example.com", "Updated", "Name", "admin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_user.test", "email", "test@example.com"),
					resource.TestCheckResourceAttr("n8n_user.test", "first_name", "Updated"),
					resource.TestCheckResourceAttr("n8n_user.test", "last_name", "Name"),
					resource.TestCheckResourceAttr("n8n_user.test", "role", "admin"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccUserResourceWithPassword(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with password
			{
				Config: testAccUserResourceConfigWithPassword("testpw@example.com", "Test", "User", "member", "testpassword123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_user.test", "email", "testpw@example.com"),
					resource.TestCheckResourceAttr("n8n_user.test", "first_name", "Test"),
					resource.TestCheckResourceAttr("n8n_user.test", "last_name", "User"),
					resource.TestCheckResourceAttr("n8n_user.test", "role", "member"),
					resource.TestCheckResourceAttrSet("n8n_user.test", "id"),
					// Password should not be in state after creation
					resource.TestCheckNoResourceAttr("n8n_user.test", "password"),
				),
			},
		},
	})
}

func TestAccUserResourceWithSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with settings
			{
				Config: testAccUserResourceConfigWithSettings("settings@example.com", "Settings", "User", "member"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_user.test", "email", "settings@example.com"),
					resource.TestCheckResourceAttr("n8n_user.test", "first_name", "Settings"),
					resource.TestCheckResourceAttr("n8n_user.test", "last_name", "User"),
					resource.TestCheckResourceAttr("n8n_user.test", "role", "member"),
					resource.TestCheckResourceAttr("n8n_user.test", "settings.theme", "dark"),
					resource.TestCheckResourceAttr("n8n_user.test", "settings.allow_sso_manual_login", "true"),
					resource.TestCheckResourceAttrSet("n8n_user.test", "id"),
				),
			},
		},
	})
}

func testAccUserResourceConfig(email, firstName, lastName, role string) string {
	return fmt.Sprintf(`
resource "n8n_user" "test" {
  email      = %[1]q
  first_name = %[2]q
  last_name  = %[3]q
  role       = %[4]q
}
`, email, firstName, lastName, role)
}

func testAccUserResourceConfigWithPassword(email, firstName, lastName, role, password string) string {
	return fmt.Sprintf(`
resource "n8n_user" "test" {
  email      = %[1]q
  first_name = %[2]q
  last_name  = %[3]q
  role       = %[4]q
  password   = %[5]q
}
`, email, firstName, lastName, role, password)
}

func testAccUserResourceConfigWithSettings(email, firstName, lastName, role string) string {
	return fmt.Sprintf(`
resource "n8n_user" "test" {
  email      = %[1]q
  first_name = %[2]q
  last_name  = %[3]q
  role       = %[4]q
  settings = {
    theme                   = "dark"
    allow_sso_manual_login = true
  }
}
`, email, firstName, lastName, role)
}
