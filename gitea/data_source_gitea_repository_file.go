package gitea

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaRepositoryFile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaRepositoryFileRead,
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Owner of the repository",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Repository name",
			},
			"file_path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Path to the file relative to repo root",
			},
			"branch": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Branch to read from; default branch if empty",
			},
			// Outputs
			"file_sha": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "File blob SHA",
			},
			"last_commit_sha": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last commit SHA touching the file",
			},
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "File size in bytes",
			},
			"content": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Base64-encoded file content",
			},
			"commit_message": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Commit message of the last commit that modified this file",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp of last commit that modified this file",
			},
		},
	}
}

func dataSourceGiteaRepositoryFileRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	owner := strings.ToLower(d.Get("username").(string))
	repo := strings.ToLower(d.Get("name").(string))
	filePath := strings.TrimPrefix(d.Get("file_path").(string), "/")
	branch := d.Get("branch").(string)

	content, resp, err := client.GetContents(owner, repo, branch, filePath)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("file not found: %s", filePath)
		}
		return err
	}

	commit, resp, err := client.GetSingleCommit(owner, repo, content.LastCommitSha)
	if err != nil {
		return err
	}
	if resp != nil && resp.StatusCode >= 400 {
		return fmt.Errorf("error getting commit: %s", resp.Status)
	}

	// Build a FileResponse-like structure to reuse state setter
	result := &gitea.FileResponse{
		Content: content,
		Commit: &gitea.FileCommitResponse{
			CommitMeta: gitea.CommitMeta{
				URL:     commit.URL,
				SHA:     commit.SHA,
				Created: resolveCommitCreatedFromRepoCommit(commit),
			},
			Author:    commit.RepoCommit.Author,
			Committer: commit.RepoCommit.Committer,
			HTMLURL:   commit.HTMLURL,
			Parents:   commit.Parents,
			Message:   commit.RepoCommit.Message,
			Tree:      commit.RepoCommit.Tree,
		},
	}

	// Set a stable ID
	d.SetId(fmt.Sprintf("%s/%s/%s/%s", owner, repo, branch, content.Path))

	// Reuse resource setter for common fields
	if err := setRepositoryFileResourceData(result, d); err != nil {
		return err
	}

	return nil
}

func resolveCommitCreatedFromRepoCommit(c *gitea.Commit) time.Time {
	if c == nil || c.RepoCommit == nil {
		return time.Time{}
	}
	if c.RepoCommit.Committer != nil {
		if t, ok := parseRFC3339Maybe(c.RepoCommit.Committer.Date); ok {
			return t
		}
	}
	if c.RepoCommit.Author != nil {
		if t, ok := parseRFC3339Maybe(c.RepoCommit.Author.Date); ok {
			return t
		}
	}
	return c.Created
}
