package gitea

import (
	"fmt"
	"os"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	testCollabOwner = "collab_acc_owner"
	testCollabUser  = "collab_acc_user"
	testCollabRepo  = "collab-acc-repo"
)

func testAccNewGiteaClient() (*gitea.Client, error) {
	baseURL := os.Getenv("GITEA_BASE_URL")
	token := os.Getenv("GITEA_TOKEN")
	return gitea.NewClient(baseURL, gitea.SetToken(token))
}

func testAccCreateCollaboratorPrerequisites(t *testing.T) {
	t.Helper()
	client, err := testAccNewGiteaClient()
	if err != nil {
		t.Fatalf("failed to create gitea client: %s", err)
	}

	pwd := "Geheim1!"
	changePassword := false

	client.AdminCreateUser(gitea.CreateUserOption{
		LoginName:          testCollabOwner,
		Username:           testCollabOwner,
		Email:              testCollabOwner + "@test.dev",
		Password:           pwd,
		MustChangePassword: &changePassword,
	})

	client.AdminCreateUser(gitea.CreateUserOption{
		LoginName:          testCollabUser,
		Username:           testCollabUser,
		Email:              testCollabUser + "@test.dev",
		Password:           pwd,
		MustChangePassword: &changePassword,
	})

	client.AdminCreateRepo(testCollabOwner, gitea.CreateRepoOption{
		Name:    testCollabRepo,
		Private: true,
	})
}

func testAccCleanupCollaboratorPrerequisites() {
	client, err := testAccNewGiteaClient()
	if err != nil {
		return
	}
	client.DeleteRepo(testCollabOwner, testCollabRepo)
	client.AdminDeleteUser(testCollabUser)
	client.AdminDeleteUser(testCollabOwner)
}

func TestAccGiteaRepositoryCollaborator_basic(t *testing.T) {
	t.Cleanup(testAccCleanupCollaboratorPrerequisites)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateCollaboratorPrerequisites(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGiteaRepositoryCollaboratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGiteaRepositoryCollaboratorConfig("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"gitea_repository_collaborator.test", "owner", testCollabOwner),
					resource.TestCheckResourceAttr(
						"gitea_repository_collaborator.test", "repo", testCollabRepo),
					resource.TestCheckResourceAttr(
						"gitea_repository_collaborator.test", "username", testCollabUser),
					resource.TestCheckResourceAttr(
						"gitea_repository_collaborator.test", "permission", "read"),
				),
			},
			{
				Config: testAccGiteaRepositoryCollaboratorConfig("write"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"gitea_repository_collaborator.test", "permission", "write"),
				),
			},
			{
				ResourceName:      "gitea_repository_collaborator.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckGiteaRepositoryCollaboratorDestroy(s *terraform.State) error {
	client, err := testAccNewGiteaClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitea_repository_collaborator" {
			continue
		}

		owner, repo, username, err := parseCollaboratorID(rs.Primary.ID)
		if err != nil {
			return err
		}

		isCollab, _, err := client.IsCollaborator(owner, repo, username)
		if err == nil && isCollab {
			return fmt.Errorf("collaborator %s still exists on %s/%s", username, owner, repo)
		}
	}

	return nil
}

func testAccGiteaRepositoryCollaboratorConfig(permission string) string {
	return fmt.Sprintf(`
resource "gitea_repository_collaborator" "test" {
  owner      = "%s"
  repo       = "%s"
  username   = "%s"
  permission = "%s"
}
`, testCollabOwner, testCollabRepo, testCollabUser, permission)
}
