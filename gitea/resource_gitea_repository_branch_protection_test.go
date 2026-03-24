package gitea

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestRepositoryBranchProtectionImporterPreservesSlashInRuleName(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryBranchProtection().Schema, map[string]interface{}{})
	d.SetId("alice/project/release/1.0/feature")

	res, err := resourceRepositoryBranchProtectionImport(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("unexpected import error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected one resource data result, got %d", len(res))
	}

	got := res[0]
	if got.Get("username").(string) != "alice" {
		t.Fatalf("expected username alice, got %v", got.Get("username"))
	}
	if got.Get("name").(string) != "project" {
		t.Fatalf("expected repo project, got %v", got.Get("name"))
	}
	if got.Get("rule_name").(string) != "release/1.0/feature" {
		t.Fatalf("expected slash-preserving rule name, got %v", got.Get("rule_name"))
	}
	if got.Id() != "release/1.0/feature" {
		t.Fatalf("expected normalized id to be the rule name, got %q", got.Id())
	}
}
