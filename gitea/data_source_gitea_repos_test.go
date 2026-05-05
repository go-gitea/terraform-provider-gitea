package gitea

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceGiteaRepos_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGiteaReposConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceGiteaReposContains(
						"data.gitea_repos.by_owner",
						"repos_data_source_test_user/repos-ds-test",
					),
					resource.TestCheckResourceAttrSet("data.gitea_repos.by_owner", "id"),
					resource.TestCheckResourceAttrSet("data.gitea_repos.by_owner", "full_names.#"),
				),
			},
		},
	})
}

// testAccDataSourceGiteaReposContains asserts that one of the entries in the
// data source's "repos" list has the expected full_name. It walks the
// flattened TypeList attributes (repos.<i>.full_name) directly so we don't
// need to know the order or count up front.
func testAccDataSourceGiteaReposContains(src, wantFullName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[src]
		if !ok {
			return fmt.Errorf("data source %s not found in state", src)
		}
		attrs := ds.Primary.Attributes

		countStr, ok := attrs["repos.#"]
		if !ok {
			return fmt.Errorf("data source %s has no repos attribute", src)
		}
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return fmt.Errorf("repos.# is not an int: %s", countStr)
		}
		if count == 0 {
			return fmt.Errorf("data source %s returned zero repos; expected to find %q", src, wantFullName)
		}

		for i := 0; i < count; i++ {
			if attrs[fmt.Sprintf("repos.%d.full_name", i)] == wantFullName {
				return nil
			}
		}
		return fmt.Errorf("data source %s does not contain repo %q (got %d repos)", src, wantFullName, count)
	}
}

func testAccDataSourceGiteaReposConfig() string {
	// The data source reads from the same Gitea instance that the provider
	// is pointed at. We materialize a known user + repo first so the
	// assertion has something deterministic to find.
	return `
resource "gitea_user" "owner" {
  username   = "repos_data_source_test_user"
  login_name = "repos_data_source_test_user"
  password   = "Geheim1!"
  email      = "repos_data_source_test_user@user.dev"
}

resource "gitea_repository" "fixture" {
  username     = gitea_user.owner.username
  name         = "repos-ds-test"
  private      = true
  issue_labels = "Default"
  license      = "MIT"
  gitignores   = "Go"
}

data "gitea_repos" "by_owner" {
  owner = gitea_user.owner.username

  depends_on = [gitea_repository.fixture]
}
`
}
