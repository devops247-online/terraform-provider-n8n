package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResource(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-test-project")
	projectDescription := "Test project description"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectResourceConfig(projectName, projectDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_project.test", "name", projectName),
					resource.TestCheckResourceAttr("n8n_project.test", "description", projectDescription),
					resource.TestCheckResourceAttrSet("n8n_project.test", "id"),
					resource.TestCheckResourceAttrSet("n8n_project.test", "created_at"),
					resource.TestCheckResourceAttrSet("n8n_project.test", "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "n8n_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccProjectResourceConfig(projectName+"-updated", projectDescription+"-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_project.test", "name", projectName+"-updated"),
					resource.TestCheckResourceAttr("n8n_project.test", "description", projectDescription+"-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccProjectResource_WithSettings(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-test-project-settings")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with settings
			{
				Config: testAccProjectResourceConfigWithSettings(projectName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_project.test", "name", projectName),
					resource.TestCheckResourceAttr("n8n_project.test", "icon", "project"),
					resource.TestCheckResourceAttr("n8n_project.test", "color", "#1f77b4"),
					resource.TestCheckResourceAttrSet("n8n_project.test", "settings"),
				),
			},
		},
	})
}

func TestAccProjectResource_MinimalConfig(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-test-project-minimal")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with minimal config
			{
				Config: testAccProjectResourceConfigMinimal(projectName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_project.test", "name", projectName),
					resource.TestCheckResourceAttrSet("n8n_project.test", "id"),
				),
			},
		},
	})
}

func testAccProjectResourceConfig(name, description string) string {
	return fmt.Sprintf(`
resource "n8n_project" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}

func testAccProjectResourceConfigWithSettings(name string) string {
	return fmt.Sprintf(`
resource "n8n_project" "test" {
  name        = %[1]q
  description = "Project with custom settings"
  icon        = "project"
  color       = "#1f77b4"
  settings    = jsonencode({
    "enableWorkflowSharing": true,
    "defaultExecutionMode": "queue"
  })
}
`, name)
}

func testAccProjectResourceConfigMinimal(name string) string {
	return fmt.Sprintf(`
resource "n8n_project" "test" {
  name = %[1]q
}
`, name)
}
