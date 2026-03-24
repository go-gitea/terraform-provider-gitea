package gitea

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	repoFilePath          = "file_path"
	encodingBase64 string = "base64"
	encodingText   string = "text"
)

// repoLocks serializes ref updates per repo/branch to avoid concurrent pushes
var repoLocks sync.Map

func repoLockKey(owner, repo, branch string) string {
	return strings.ToLower(fmt.Sprintf("%s/%s/%s", owner, repo, branch))
}

func getRepoMutex(owner, repo, branch string) *sync.Mutex {
	key := repoLockKey(owner, repo, branch)
	m, _ := repoLocks.LoadOrStore(key, &sync.Mutex{})
	return m.(*sync.Mutex)
}

func buildRepositoryFileID(owner, repo, branch, filePath string) string {
	return fmt.Sprintf("%s/%s/%s/%s", owner, repo, url.QueryEscape(branch), url.QueryEscape(filePath))
}

func parseRepositoryFileID(id string) (owner, repo, branch, filePath string, err error) {
	parts := strings.SplitN(id, "/", 4)
	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("unexpected ID format (%q), expected <username>/<repo>/<branch>/<file_path>", id)
	}

	owner = parts[0]
	repo = parts[1]
	branch, err = url.QueryUnescape(parts[2])
	if err != nil {
		return "", "", "", "", fmt.Errorf("invalid branch in ID %q: %w", id, err)
	}
	filePath, err = url.QueryUnescape(parts[3])
	if err != nil {
		return "", "", "", "", fmt.Errorf("invalid file path in ID %q: %w", id, err)
	}

	return owner, repo, branch, filePath, nil
}

func resourceRepositoryFileRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	username, name, branch, filePath, err := parseRepositoryFileID(d.Id())
	if err != nil {
		return err
	}

	if err := d.Set("username", username); err != nil {
		return err
	}
	if err := d.Set("name", name); err != nil {
		return err
	}
	if err := d.Set("branch", branch); err != nil {
		return err
	}
	if err := d.Set("file_path", filePath); err != nil {
		return err
	}

	// Get current file metadata and contents
	content, resp, err := client.GetContents(username, name, branch, filePath)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}

	result := &gitea.FileResponse{
		Content: content,
	}
	if content.LastCommitSha != nil && *content.LastCommitSha != "" {
		commit, resp, err := client.GetSingleCommit(username, name, *content.LastCommitSha)
		if err != nil {
			return err
		}
		if resp != nil && resp.StatusCode != 200 {
			return fmt.Errorf("error getting commit from repository: %s", resp.Status)
		}

		// Prefer committer/author date; CommitMeta.Created is often zero
		created := time.Time{}
		if commit.RepoCommit != nil {
			if commit.RepoCommit.Committer != nil {
				if t, ok := parseRFC3339Maybe(commit.RepoCommit.Committer.Date); ok {
					created = t
				}
			}
			if created.IsZero() && commit.RepoCommit.Author != nil {
				if t, ok := parseRFC3339Maybe(commit.RepoCommit.Author.Date); ok {
					created = t
				}
			}
		}
		if created.IsZero() {
			created = commit.Created
		}

		result.Commit = &gitea.FileCommitResponse{
			CommitMeta: gitea.CommitMeta{
				URL:     commit.URL,
				SHA:     commit.SHA,
				Created: created,
			},
			Author:    commit.RepoCommit.Author,
			Committer: commit.RepoCommit.Committer,
			HTMLURL:   commit.HTMLURL,
			Parents:   commit.Parents,
			Message:   commit.RepoCommit.Message,
			Tree:      commit.RepoCommit.Tree,
		}
	}
	err = setRepositoryFileResourceData(result, d)

	return
}
func resourceRepositoryFileCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	usernameData, usernameOk := d.GetOk("username")
	if !usernameOk {
		return fmt.Errorf("name of repo owner must be passed")
	}
	username := strings.ToLower(usernameData.(string))

	nameData, nameOk := d.GetOk("name")
	if !nameOk {
		return fmt.Errorf("name of repo must be passed")
	}
	name := strings.ToLower(nameData.(string))

	filePathData, filePathOk := d.GetOk("file_path")
	if !filePathOk {
		return fmt.Errorf("file path must be passed")
	}
	filePath := strings.TrimPrefix(filePathData.(string), "/")

	branch := d.Get("branch").(string)

	commitMessage := d.Get("commit_message").(string)

	overwrite := d.Get("overwrite").(bool)

	fileContentData, fileContentOk := d.GetOk("content")
	if !fileContentOk {
		return fmt.Errorf("file content must be passed")
	}
	fileContent := fileContentData.(string)

	encoding := d.Get("encoding")
	switch encoding.(string) {
	case encodingText:
		fileContent = base64.StdEncoding.EncodeToString([]byte(fileContent))
	case encodingBase64:
		// fileContent is already base64 encoded
	default:
		return fmt.Errorf("encoding must be one of 'base64' or 'text'")
	}

	// Lock per repo/branch to prevent concurrent ref updates
	getRepoMutex(username, name, branch).Lock()
	defer getRepoMutex(username, name, branch).Unlock()

	// Check if the file exists first; if 404, we'll create it; other errors bubble up
	content, resp, err := client.GetContents(username, name, branch, filePath)
	exists := false
	if err != nil {
		// If resp is available and indicates not found, treat as non-existent; otherwise return error
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			exists = false
		} else {
			return fmt.Errorf("error checking file existence: %w", err)
		}
	} else if resp != nil && resp.StatusCode == http.StatusOK {
		exists = true
	}

	var fileResponse *gitea.FileResponse
	if exists {
		if !overwrite {
			return fmt.Errorf("file already exists and overwrite is not allowed")
		}
		// File exists, update it
		updateOpts := gitea.UpdateFileOptions{
			Content: fileContent,
			SHA:     content.SHA,
			FileOptions: gitea.FileOptions{
				Message:    commitMessage,
				BranchName: branch,
			},
		}
		// Retry on concurrent branch updates
		for attempts, backoff := 0, 200*time.Millisecond; ; attempts, backoff = attempts+1, backoff*2 {
			var resp *gitea.Response
			fileResponse, resp, err = client.UpdateFile(username, name, filePath, updateOpts)
			if err == nil {
				break
			}
			if shouldRetryPush(resp, err) && attempts < 5 {
				// refresh SHA in case the branch advanced
				if fresh, _, gerr := client.GetContents(username, name, branch, filePath); gerr == nil && fresh != nil {
					updateOpts.SHA = fresh.SHA
				}
				time.Sleep(backoff)
				continue
			}
			return fmt.Errorf("error updating file in repository: %v", err)
		}
	} else {
		// File does not exist, create it
		createOpts := gitea.CreateFileOptions{
			Content: fileContent,
			FileOptions: gitea.FileOptions{
				Message:    commitMessage,
				BranchName: branch,
			},
		}
		// Retry on concurrent branch updates
		for attempts, backoff := 0, 200*time.Millisecond; ; attempts, backoff = attempts+1, backoff*2 {
			var resp *gitea.Response
			fileResponse, resp, err = client.CreateFile(username, name, filePath, createOpts)
			if err == nil {
				break
			}
			if shouldRetryPush(resp, err) && attempts < 5 {
				time.Sleep(backoff)
				continue
			}
			return fmt.Errorf("error creating file in repository: %v", err)
		}
	}

	err = setRepositoryFileResourceData(fileResponse, d)

	return
}

func resourceRepositoryFileUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	name := d.Get(repoName).(string)
	username := d.Get(repoOwner).(string)
	filePath := d.Get(repoFilePath).(string)
	branch := d.Get("branch").(string)
	content := d.Get("content").(string)
	commitMessage := d.Get("commit_message").(string)
	// last_commit_sha := d.Get("last_commit_sha").(string)
	file_sha := d.Get("file_sha").(string)

	encoding := d.Get("encoding")
	switch encoding.(string) {
	case encodingText:
		content = base64.StdEncoding.EncodeToString([]byte(content))
	case encodingBase64:
		// Content is already base64 encoded
	default:
		return fmt.Errorf("encoding must be one of 'base64' or 'text'")
	}
	opts := gitea.UpdateFileOptions{
		FileOptions: gitea.FileOptions{
			Message:    commitMessage,
			BranchName: branch,
		},
		SHA:     file_sha,
		Content: content,
	}

	// Lock per repo/branch to prevent concurrent ref updates
	getRepoMutex(username, name, branch).Lock()
	defer getRepoMutex(username, name, branch).Unlock()

	var fileResponse *gitea.FileResponse
	var resp *gitea.Response
	for attempts, backoff := 0, 200*time.Millisecond; ; attempts, backoff = attempts+1, backoff*2 {
		fileResponse, resp, err = client.UpdateFile(username, name, filePath, opts)
		if err == nil {
			break
		}
		if shouldRetryPush(resp, err) && attempts < 5 {
			// refresh SHA before retry
			if fresh, _, gerr := client.GetContents(username, name, branch, filePath); gerr == nil && fresh != nil {
				opts.SHA = fresh.SHA
			}
			time.Sleep(backoff)
			continue
		}
		return fmt.Errorf("error updating file in repository: %v", err)
	}
	// File exists, update it
	// fileResponse, resp, err := client.UpdateFile(username, name, filePath, opts)
	if err != nil {
		return err
	}

	err = setRepositoryFileResourceData(fileResponse, d)
	if err != nil {
		return fmt.Errorf("error setting file resource data: %v", err)
	}

	return nil

}

func resourceRepositoryFileDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	owner := d.Get(repoOwner).(string)
	name := d.Get(repoName).(string)
	filePath := d.Get(repoFilePath).(string)
	fileSHA := d.Get("file_sha").(string)
	branch := d.Get("branch").(string)

	// Lock per repo/branch to prevent concurrent ref updates
	getRepoMutex(owner, name, branch).Lock()
	defer getRepoMutex(owner, name, branch).Unlock()

	var resp *gitea.Response
	for attempts, backoff := 0, 200*time.Millisecond; ; attempts, backoff = attempts+1, backoff*2 {
		resp, err = client.DeleteFile(owner, name, filePath, gitea.DeleteFileOptions{
			SHA: fileSHA,
			FileOptions: gitea.FileOptions{
				BranchName: branch,
			},
		})
		if err == nil {
			break
		}
		if shouldRetryPush(resp, err) && attempts < 5 {
			time.Sleep(backoff)
			continue
		}
		return fmt.Errorf("error deleting file from repository: %v", err)
	}

	return nil
}

func setRepositoryFileResourceData(response *gitea.FileResponse, d *schema.ResourceData) (err error) {
	// Make a unique ID for the resource from the repo and file path
	d.Set("file_path", response.Content.Path)
	d.SetId(buildRepositoryFileID(
		d.Get("username").(string),
		d.Get("name").(string),
		d.Get("branch").(string),
		response.Content.Path,
	))
	d.Set("file_sha", response.Content.SHA)
	lastCommitSHA := ""
	if response.Commit != nil {
		lastCommitSHA = response.Commit.SHA
	}
	d.Set("last_commit_sha", lastCommitSHA)
	d.Set("size", response.Content.Size)
	// Preserve the user-provided commit_message in state to avoid perpetual diffs
	if v, ok := d.GetOk("commit_message"); ok {
		d.Set("commit_message", v.(string))
	} else {
		d.Set("commit_message", "")
	}
	d.Set("content", response.Content.Content) // This is base64 encoded content
	// Resolve created_at from commit fields without emitting zero time
	created := resolveCommitCreated(response.Commit)
	if !created.IsZero() {
		d.Set("created_at", created.Format(time.RFC3339))
	} else {
		d.Set("created_at", "")
	}

	return
}

// resolveCommitCreated prefers committer/author date; falls back to CommitMeta.Created
func resolveCommitCreated(c *gitea.FileCommitResponse) time.Time {
	if c == nil {
		return time.Time{}
	}
	if c.Committer != nil {
		if t, ok := parseRFC3339Maybe(c.Committer.Date); ok {
			return t
		}
	}
	if c.Author != nil {
		if t, ok := parseRFC3339Maybe(c.Author.Date); ok {
			return t
		}
	}
	return c.CommitMeta.Created
}

// parseRFC3339Maybe attempts to parse common RFC3339 formats
func parseRFC3339Maybe(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	// Try RFC3339Nano first, then RFC3339
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}
	return time.Time{}, false
}

// shouldRetryPush returns true for errors that look like concurrent ref updates
// from Gitea where the server expects a different tip (HTTP 409 or PushRejected wording).
func shouldRetryPush(resp *gitea.Response, err error) bool {
	if resp != nil && resp.StatusCode == http.StatusConflict { // 409
		return true
	}
	if err == nil {
		return false
	}
	// Match common error text returned by git backend in Gitea
	msg := err.Error()
	// e.g. "PushRejected Error: exit status 1 - remote: error: cannot lock ref 'refs/heads/main': is at ... but expected ..."
	if strings.Contains(msg, "PushRejected") || strings.Contains(msg, "cannot lock ref") {
		return true
	}
	// conservative: retry on generic "failed to push some refs" once
	re := regexp.MustCompile(`failed to push some refs`)
	return re.MatchString(msg)
}

func resourceGiteaRepositoryFile() *schema.Resource {
	return &schema.Resource{
		Read:   resourceRepositoryFileRead,
		Create: resourceRepositoryFileCreate,
		Update: resourceRepositoryFileUpdate,
		Delete: resourceRepositoryFileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Owner of the repository",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Name of the repository",
			},
			"file_path": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The path to the file relative to the root of the repository",
			},
			"encoding": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					if v, ok := val.(string); ok {
						if v != "base64" && v != "text" {
							errs = append(errs, fmt.Errorf("Invalid encoding type: %s. Allowed values are 'base64' and 'text'", v))
						}
					}
					return
				},
				Default:     "text",
				Description: "The encoding of the file content. Currently only 'base64' and 'text' are supported. Defaults to 'text'",
			},
			"content": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "The content of the file",
				// The content from the API is always base64 encoded, so we might need to decode it to see if it has changed
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {

					var oldDecoded string

					oldDecodedBytes, err := base64.StdEncoding.DecodeString(old)
					if err != nil {
						oldDecoded = old // treat as raw text
					} else {
						oldDecoded = string(oldDecodedBytes)
					}
					return oldDecoded == new
				},
			},
			"branch": {
				Type:        schema.TypeString,
				Required:    false,
				ForceNew:    false,
				Optional:    true,
				Default:     "",
				Description: "The branch to create or modify the file in. If not provided, the default branch will be used",
			},
			"overwrite": {
				Type:        schema.TypeBool,
				Required:    false,
				ForceNew:    false,
				Optional:    true,
				Default:     true,
				Description: "Whether to overwrite the file if it already exists. This is only applicable when creating a new file.",
			},
			"commit_message": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				// Write-only input: ignore diffs to avoid unnecessary updates
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool { return true },
				Description:      "The commit message to use when creating or updating the file. If not supplied, a default message will be used",
			},
			"file_sha": {
				Type:        schema.TypeString,
				Required:    false,
				ForceNew:    false,
				Computed:    true,
				Description: "The SHA of the file",
			},
			"last_commit_sha": {
				Type:        schema.TypeString,
				Required:    false,
				ForceNew:    false,
				Computed:    true,
				Description: "The SHA of the last commit",
			},
			"size": {
				Type:        schema.TypeInt,
				Required:    false,
				ForceNew:    false,
				Computed:    true,
				Description: "The size of the file in bytes",
			},
			"created_at": {
				Type:        schema.TypeString,
				Required:    false,
				ForceNew:    false,
				Computed:    true,
				Description: "The creation date of the file commit",
			},
		},
		Description: "`gitea_repository_file` manages a file in a gitea repository.\n\n" +
			"If the file does not exist it will be created. If the file exists and overwrite is false, an error is returned.\n",
	}
}
