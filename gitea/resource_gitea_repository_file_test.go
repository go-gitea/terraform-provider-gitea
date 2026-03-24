package gitea

import (
	"encoding/base64"
	"net/url"
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestRepositoryFileIDRoundTrip(t *testing.T) {
	owner := "alice"
	repo := "project"
	branch := "feature/foo bar"
	filePath := "docs/read me.txt"

	id := buildRepositoryFileID(owner, repo, branch, filePath)
	wantID := owner + "/" + repo + "/" + url.QueryEscape(branch) + "/" + url.QueryEscape(filePath)
	if id != wantID {
		t.Fatalf("expected canonical id %q, got %q", wantID, id)
	}

	gotOwner, gotRepo, gotBranch, gotFilePath, err := parseRepositoryFileID(id)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if gotOwner != owner || gotRepo != repo || gotBranch != branch || gotFilePath != filePath {
		t.Fatalf("round trip mismatch: got %q/%q/%q/%q", gotOwner, gotRepo, gotBranch, gotFilePath)
	}
}

func TestRepositoryFileParseLegacyIDBestEffort(t *testing.T) {
	owner, repo, branch, filePath, err := parseRepositoryFileID("alice/project/main/docs/readme.md")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if owner != "alice" || repo != "project" || branch != "main" || filePath != "docs/readme.md" {
		t.Fatalf("unexpected legacy parse result: %q/%q/%q/%q", owner, repo, branch, filePath)
	}
}

func TestSetRepositoryFileResourceDataUsesCanonicalID(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryFile().Schema, map[string]interface{}{
		"username":       "alice",
		"name":           "project",
		"branch":         "feature/foo",
		"file_path":      "docs/read me.txt",
		"commit_message": "keep",
	})

	content := base64.StdEncoding.EncodeToString([]byte("hello"))
	response := &gitea.FileResponse{
		Content: &gitea.ContentsResponse{
			Path:    "docs/read me.txt",
			SHA:     "blob-sha",
			Size:    5,
			Content: &content,
		},
		Commit: &gitea.FileCommitResponse{
			CommitMeta: gitea.CommitMeta{
				SHA:     "commit-sha",
				Created: time.Unix(10, 0),
			},
		},
	}

	if err := setRepositoryFileResourceData(response, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantID := buildRepositoryFileID("alice", "project", "feature/foo", "docs/read me.txt")
	if d.Id() != wantID {
		t.Fatalf("expected canonical id %q, got %q", wantID, d.Id())
	}
	if got := d.Get("file_path").(string); got != "docs/read me.txt" {
		t.Fatalf("expected file_path to be authoritative, got %q", got)
	}
}
