package gitea

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRepoBranchIdPartsPreservesSlashInBranchName(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryBranch().Schema, map[string]interface{}{})
	d.SetId("123/feature/foo/bar")

	hasID, repoID, branchID, err := resourceRepoBranchIdParts(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasID {
		t.Fatalf("expected ID to parse")
	}
	if repoID != 123 {
		t.Fatalf("expected repo ID 123, got %d", repoID)
	}
	if branchID != "feature/foo/bar" {
		t.Fatalf("expected branch name to be preserved, got %q", branchID)
	}
}
