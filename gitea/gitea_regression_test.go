package gitea

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestSetRepositoryFileResourceDataAllowsNilCommit(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryFile().Schema, map[string]interface{}{
		"username":  "alice",
		"name":      "project",
		"branch":    "main",
		"file_path": "README.md",
	})

	content := "aGVsbG8="
	if err := setRepositoryFileResourceData(&gitea.FileResponse{
		Content: &gitea.ContentsResponse{
			Path:    "README.md",
			SHA:     "blob-sha",
			Size:    5,
			Content: &content,
		},
	}, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := d.Get(lastCommitSHAField).(string); got != "" {
		t.Fatalf("expected empty last_commit_sha, got %q", got)
	}
}

func TestResolveActionScope(t *testing.T) {
	tests := []struct {
		name      string
		values    map[string]interface{}
		wantScope actionScopeConfig
		wantErr   bool
	}{
		{
			name:      "admin",
			values:    map[string]interface{}{actionScopeField: actionScopeAdmin},
			wantScope: actionScopeConfig{Scope: actionScopeAdmin},
		},
		{
			name:      "user",
			values:    map[string]interface{}{actionScopeField: actionScopeUser},
			wantScope: actionScopeConfig{Scope: actionScopeUser},
		},
		{
			name:      "org",
			values:    map[string]interface{}{actionScopeField: actionScopeOrg, actionOrgField: "example"},
			wantScope: actionScopeConfig{Scope: actionScopeOrg, Org: "example"},
		},
		{
			name:      "repo",
			values:    map[string]interface{}{actionScopeField: actionScopeRepo, repositoryOwnerField: "alice", repositoryNameField: "demo"},
			wantScope: actionScopeConfig{Scope: actionScopeRepo, Owner: "alice", Repo: "demo"},
		},
		{
			name:    "missing repo fields",
			values:  map[string]interface{}{actionScopeField: actionScopeRepo},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, mergeSchemaMaps(actionScopeSchema(), map[string]*schema.Schema{}), tc.values)
			got, err := resolveActionScope(d)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantScope {
				t.Fatalf("unexpected scope config: %#v", got)
			}
		})
	}
}

func TestOptionalBoolValuePreservesExplicitFalse(t *testing.T) {
	s := map[string]*schema.Schema{
		"disabled": {
			Type:     schema.TypeBool,
			Optional: true,
		},
	}

	t.Run("unset", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, s, map[string]interface{}{})
		if got := optionalBoolValue(d, "disabled"); got != nil {
			t.Fatalf("expected nil pointer, got %v", *got)
		}
	})

	t.Run("true", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, s, map[string]interface{}{"disabled": true})
		got := optionalBoolValue(d, "disabled")
		if got == nil || !*got {
			t.Fatalf("expected true pointer, got %#v", got)
		}
	})

	t.Run("false", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, s, map[string]interface{}{"disabled": false})
		got := optionalBoolValue(d, "disabled")
		if got == nil || *got {
			t.Fatalf("expected false pointer, got %#v", got)
		}
	})
}

func TestCollectPaginatedWithLimitErrorsWhenSafetyLimitReached(t *testing.T) {
	_, err := collectPaginatedWithLimit(2, func(page int) ([]int, error) {
		return []int{page}, nil
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "pagination exceeded safety limit") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCollectPaginatedWithLimitReturnsAllItemsBeforeEmptyPage(t *testing.T) {
	items, err := collectPaginatedWithLimit(3, func(page int) ([]int, error) {
		if page == 3 {
			return []int{}, nil
		}
		return []int{page}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0] != 1 || items[1] != 2 {
		t.Fatalf("unexpected items: %#v", items)
	}
}

func TestCollectPaginatedWithLimitAllowsExactLimitWhenNextPageIsEmpty(t *testing.T) {
	items, err := collectPaginatedWithLimit(2, func(page int) ([]int, error) {
		if page == 3 {
			return []int{}, nil
		}
		return []int{page}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0] != 1 || items[1] != 2 {
		t.Fatalf("unexpected items: %#v", items)
	}
}

func TestIssueAttachmentImporterParsesCompositeID(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaIssueAttachment().Schema, map[string]interface{}{})
	d.SetId("alice:demo:12:99")

	out, err := resourceGiteaIssueAttachment().Importer.StateContext(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected one resource, got %d", len(out))
	}
	if got := out[0].Get(repositoryOwnerField).(string); got != "alice" {
		t.Fatalf("unexpected repository owner %q", got)
	}
	if got := out[0].Get(repositoryNameField).(string); got != "demo" {
		t.Fatalf("unexpected repository %q", got)
	}
	if got := out[0].Get("issue_index").(int); got != 12 {
		t.Fatalf("unexpected issue index %d", got)
	}
}

func TestIssueAttachmentImporterRejectsInvalidAttachmentID(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaIssueAttachment().Schema, map[string]interface{}{})
	d.SetId("alice:demo:12:not-a-number")

	_, err := resourceGiteaIssueAttachment().Importer.StateContext(context.Background(), d, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRepositoryActionsWorkflowStateImporterParsesCompositeID(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryActionsWorkflowState().Schema, map[string]interface{}{})
	d.SetId("alice:demo:ci.yml")

	out, err := resourceGiteaRepositoryActionsWorkflowState().Importer.StateContext(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected one resource, got %d", len(out))
	}
	if got := out[0].Get(repositoryOwnerField).(string); got != "alice" {
		t.Fatalf("unexpected repository owner %q", got)
	}
	if got := out[0].Get(repositoryNameField).(string); got != "demo" {
		t.Fatalf("unexpected repository %q", got)
	}
	if got := out[0].Get(workflowIDField).(string); got != "ci.yml" {
		t.Fatalf("unexpected workflow id %q", got)
	}
}

func TestRepositoryActionsWorkflowStateUsesDedicatedDeleteHandler(t *testing.T) {
	resource := resourceGiteaRepositoryActionsWorkflowState()
	if resource.Delete == nil {
		t.Fatal("expected delete function")
	}
	if reflect.ValueOf(resource.Delete).Pointer() != reflect.ValueOf(resourceGiteaRepositoryActionsWorkflowStateDelete).Pointer() {
		t.Fatal("expected workflow state resource to use dedicated delete handler")
	}
}
