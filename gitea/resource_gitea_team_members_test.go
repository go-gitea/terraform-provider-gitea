package gitea

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestTeamMembersDiffDetectsAddsAndRemovals(t *testing.T) {
	toAdd, toRemove := teamMembersDiff(
		[]string{"alice", "bob", "carol"},
		[]string{"bob", "carol", "dave"},
	)

	if len(toAdd) != 1 || toAdd[0] != "dave" {
		t.Fatalf("expected dave to be added, got %#v", toAdd)
	}
	if len(toRemove) != 1 || toRemove[0] != "alice" {
		t.Fatalf("expected alice to be removed, got %#v", toRemove)
	}
}

func TestTeamMembersImporterParsesTeamID(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaTeamMembers().Schema, map[string]interface{}{})
	d.SetId("17")

	res, err := resourceGiteaTeamMembers().Importer.StateContext(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("unexpected import error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected one resource data result, got %d", len(res))
	}

	got := res[0]
	if got.Id() != "17" {
		t.Fatalf("expected normalized id 17, got %q", got.Id())
	}
	if got.Get("team_id").(int) != 17 {
		t.Fatalf("expected team_id 17, got %v", got.Get("team_id"))
	}
}
