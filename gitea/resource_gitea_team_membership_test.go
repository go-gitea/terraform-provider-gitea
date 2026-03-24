package gitea

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestParseTeamMembershipIDSupportsCanonicalAndLegacyFormats(t *testing.T) {
	teamID, username, err := parseTeamMembershipID("42/alice")
	if err != nil {
		t.Fatalf("unexpected canonical parse error: %v", err)
	}
	if teamID != 42 || username != "alice" {
		t.Fatalf("unexpected canonical parse result: %d/%q", teamID, username)
	}

	teamID, username, err = parseTeamMembershipID("42_alice")
	if err != nil {
		t.Fatalf("unexpected legacy parse error: %v", err)
	}
	if teamID != 42 || username != "alice" {
		t.Fatalf("unexpected legacy parse result: %d/%q", teamID, username)
	}
}

func TestTeamMembershipImporterParsesSlashID(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaTeamMembership().Schema, map[string]interface{}{})
	d.SetId("42/alice")

	res, err := resourceGiteaTeamMembership().Importer.StateContext(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("unexpected import error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected one resource data result, got %d", len(res))
	}

	got := res[0]
	if got.Id() != "42/alice" {
		t.Fatalf("expected normalized id %q, got %q", "42/alice", got.Id())
	}
	if got.Get("team_id").(int) != 42 {
		t.Fatalf("expected team_id 42, got %v", got.Get("team_id"))
	}
	if got.Get("username").(string) != "alice" {
		t.Fatalf("expected username alice, got %v", got.Get("username"))
	}
}
