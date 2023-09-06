package gitea

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceGiteaUser_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Get user using its username
			{
				Config: testAccDataGiteaUserConfigUsername(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceGiteaUser("data.gitea_user.foo"),
				),
			},
			{
				Config: testAccDataGiteaUserConfigUsername(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceGiteaUser("data.gitea_user.self"),
				),
			},
		},
	})
}

func testAccDataSourceGiteaUser(src string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		user := s.RootModule().Resources[src]
		userResource := user.Primary.Attributes

		testAttributes := []string{
			"username",
		}

		for _, attribute := range testAttributes {
			if userResource[attribute] != "test01" {
				return fmt.Errorf("Expected user's parameter `%s` to be: %s, but got: `%s`", attribute, userResource[attribute], "test01")
			}
		}

		return nil
	}
}

func testAccDataGiteaUserConfigUsername() string {
	return fmt.Sprintf(`
data "gitea_user" "foo" {
  username = "test01"
}
data "gitea_user" "self" {
}
`)
}
