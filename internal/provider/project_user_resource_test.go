package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectUserResource(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-test-project")
	userEmail := fmt.Sprintf("test-%s@example.com", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectUserResourceConfig(projectName, userEmail, "editor"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_project_user.test", "user_id", userEmail),
					resource.TestCheckResourceAttr("n8n_project_user.test", "role", "editor"),
					resource.TestCheckResourceAttrSet("n8n_project_user.test", "id"),
					resource.TestCheckResourceAttrSet("n8n_project_user.test", "project_id"),
					resource.TestCheckResourceAttrSet("n8n_project_user.test", "added_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "n8n_project_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing (role change)
			{
				Config: testAccProjectUserResourceConfig(projectName, userEmail, "admin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_project_user.test", "role", "admin"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccProjectUserResource_DefaultRole(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-test-project")
	userEmail := fmt.Sprintf("test-%s@example.com", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with default role
			{
				Config: testAccProjectUserResourceConfigDefaultRole(projectName, userEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_project_user.test", "user_id", userEmail),
					resource.TestCheckResourceAttr("n8n_project_user.test", "role", "viewer"),
				),
			},
		},
	})
}

func TestAccProjectUserResource_MultipleUsers(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-test-project")
	userEmail1 := fmt.Sprintf("test1-%s@example.com", acctest.RandString(8))
	userEmail2 := fmt.Sprintf("test2-%s@example.com", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with multiple users
			{
				Config: testAccProjectUserResourceConfigMultiple(projectName, userEmail1, userEmail2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_project_user.test1", "user_id", userEmail1),
					resource.TestCheckResourceAttr("n8n_project_user.test1", "role", "admin"),
					resource.TestCheckResourceAttr("n8n_project_user.test2", "user_id", userEmail2),
					resource.TestCheckResourceAttr("n8n_project_user.test2", "role", "editor"),
				),
			},
		},
	})
}

func testAccProjectUserResourceConfig(projectName, userEmail, role string) string {
	return fmt.Sprintf(`
resource "n8n_project" "test" {
  name        = %[1]q
  description = "Test project for user assignment"
}

resource "n8n_user" "test" {
  email      = %[2]q
  first_name = "Test"
  last_name  = "User"
  password   = "TempPassword123!"
}

resource "n8n_project_user" "test" {
  project_id = n8n_project.test.id
  user_id    = n8n_user.test.id
  role       = %[3]q
}
`, projectName, userEmail, role)
}

func testAccProjectUserResourceConfigDefaultRole(projectName, userEmail string) string {
	return fmt.Sprintf(`
resource "n8n_project" "test" {
  name        = %[1]q
  description = "Test project for user assignment"
}

resource "n8n_user" "test" {
  email      = %[2]q
  first_name = "Test"
  last_name  = "User"
  password   = "TempPassword123!"
}

resource "n8n_project_user" "test" {
  project_id = n8n_project.test.id
  user_id    = n8n_user.test.id
}
`, projectName, userEmail)
}

func testAccProjectUserResourceConfigMultiple(projectName, userEmail1, userEmail2 string) string {
	return fmt.Sprintf(`
resource "n8n_project" "test" {
  name        = %[1]q
  description = "Test project for multiple user assignment"
}

resource "n8n_user" "test1" {
  email      = %[2]q
  first_name = "Test1"
  last_name  = "User"
  password   = "TempPassword123!"
}

resource "n8n_user" "test2" {
  email      = %[3]q
  first_name = "Test2"
  last_name  = "User"
  password   = "TempPassword123!"
}

resource "n8n_project_user" "test1" {
  project_id = n8n_project.test.id
  user_id    = n8n_user.test1.id
  role       = "admin"
}

resource "n8n_project_user" "test2" {
  project_id = n8n_project.test.id
  user_id    = n8n_user.test2.id
  role       = "editor"
}
`, projectName, userEmail1, userEmail2)
}
