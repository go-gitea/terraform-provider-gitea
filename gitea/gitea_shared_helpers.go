package gitea

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	actionScopeField     = "scope"
	actionOrgField       = "org"
	sourcePathField      = "source_path"
	descriptionField     = "description"
	repositoryOwnerField = "repository_owner"
	repositoryNameField  = "repository"
	workflowIDField      = "workflow_id"
	runIDField           = "run_id"
	enabledField         = "enabled"
	createdAtField       = "created_at"
	updatedAtField       = "updated_at"
	lastCommitSHAField   = "last_commit_sha"
)

const (
	actionScopeAdmin = "admin"
	actionScopeUser  = "user"
	actionScopeOrg   = "org"
	actionScopeRepo  = "repo"
)

type actionScopeConfig struct {
	Scope string
	Org   string
	Owner string
	Repo  string
}

func actionScopeSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		actionScopeField: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The Actions scope: `admin`, `user`, `org`, or `repo`.",
		},
		actionOrgField: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "The organisation name when `scope = \"org\"`.",
		},
		repositoryOwnerField: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "The repository owner when `scope = \"repo\"`.",
		},
		repositoryNameField: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "The repository name when `scope = \"repo\"`.",
		},
	}
}

func mergeSchemaMaps(maps ...map[string]*schema.Schema) map[string]*schema.Schema {
	merged := make(map[string]*schema.Schema)
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

func repositoryIdentitySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		repositoryOwnerField: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The repository owner.",
		},
		repositoryNameField: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The repository name.",
		},
	}
}

func resolveActionScope(d *schema.ResourceData) (actionScopeConfig, error) {
	scope := strings.ToLower(strings.TrimSpace(d.Get(actionScopeField).(string)))
	cfg := actionScopeConfig{
		Scope: scope,
		Org:   strings.TrimSpace(d.Get(actionOrgField).(string)),
		Owner: strings.TrimSpace(d.Get(repositoryOwnerField).(string)),
		Repo:  strings.TrimSpace(d.Get(repositoryNameField).(string)),
	}

	switch scope {
	case actionScopeAdmin, actionScopeUser:
		if cfg.Org != "" || cfg.Owner != "" || cfg.Repo != "" {
			return cfg, fmt.Errorf("scope %q does not accept org or repository fields", scope)
		}
	case actionScopeOrg:
		if cfg.Org == "" {
			return cfg, fmt.Errorf("org must be set when scope = %q", actionScopeOrg)
		}
		if cfg.Owner != "" || cfg.Repo != "" {
			return cfg, fmt.Errorf("scope %q does not accept repository fields", actionScopeOrg)
		}
	case actionScopeRepo:
		if cfg.Owner == "" || cfg.Repo == "" {
			return cfg, fmt.Errorf("repository_owner and repository must be set when scope = %q", actionScopeRepo)
		}
		if cfg.Org != "" {
			return cfg, fmt.Errorf("scope %q does not accept org", actionScopeRepo)
		}
	default:
		return cfg, fmt.Errorf("unsupported scope %q", scope)
	}

	return cfg, nil
}

func versionError(feature string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", feature, err)
}

func requireVersion(client *gitea.Client, constraint, feature string) error {
	return versionError(feature, client.CheckServerVersionConstraint(constraint))
}

func readLocalFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("path must be set")
	}
	return os.ReadFile(filepath.Clean(path))
}

func basenameOrValue(path, fallback string) string {
	if fallback != "" {
		return fallback
	}
	return filepath.Base(path)
}

func optionalBoolValue(d *schema.ResourceData, field string) *bool {
	if raw, ok := d.GetOkExists(field); ok {
		value := raw.(bool)
		return &value
	}

	return nil
}

func buildResourceID(parts ...string) string {
	return strings.Join(parts, ":")
}

func buildTwoPartID(left, right string) string {
	return fmt.Sprintf("%s:%s", left, right)
}

func parseTwoPartID(id, leftName, rightName string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected ID format (%q). Expected %s:%s", id, leftName, rightName)
	}
	return parts[0], parts[1], nil
}

func buildThreePartID(a, b, c string) string {
	return fmt.Sprintf("%s:%s:%s", a, b, c)
}

func parseThreePartID(id, left, center, right string) (string, string, string, error) {
	parts := strings.SplitN(id, ":", 3)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("unexpected ID format (%q). Expected %s:%s:%s", id, left, center, right)
	}
	return parts[0], parts[1], parts[2], nil
}

func buildFourPartID(a, b, c, d string) string {
	return fmt.Sprintf("%s:%s:%s:%s", a, b, c, d)
}

func parseFourPartID(id, aName, bName, cName, dName string) (string, string, string, string, error) {
	parts := strings.SplitN(id, ":", 4)
	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("unexpected ID format (%q). Expected %s:%s:%s:%s", id, aName, bName, cName, dName)
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

func flattenUsers(users []*gitea.User) []interface{} {
	if len(users) == 0 {
		return []interface{}{}
	}
	result := make([]interface{}, 0, len(users))
	for _, user := range users {
		if user == nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"id":         int(user.ID),
			"username":   user.UserName,
			"full_name":  user.FullName,
			"email":      user.Email,
			"avatar_url": user.AvatarURL,
		})
	}
	return result
}

func flattenContactLinks(links []gitea.IssueConfigContactLink) []interface{} {
	if len(links) == 0 {
		return []interface{}{}
	}
	result := make([]interface{}, 0, len(links))
	for _, link := range links {
		result = append(result, map[string]interface{}{
			"name":  link.Name,
			"url":   link.URL,
			"about": link.About,
		})
	}
	return result
}

func flattenAttachments(items []*gitea.Attachment) []interface{} {
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"id":             item.ID,
			"name":           item.Name,
			"size":           item.Size,
			"download_count": item.DownloadCount,
			"created_at":     item.Created.Format(time.RFC3339),
			"uuid":           item.UUID,
			"download_url":   item.DownloadURL,
		})
	}
	return result
}

func flattenActionRunners(items []*gitea.ActionRunner) []interface{} {
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		labels := make([]interface{}, 0, len(item.Labels))
		for _, label := range item.Labels {
			if label == nil {
				continue
			}
			labels = append(labels, map[string]interface{}{
				"id":   label.ID,
				"name": label.Name,
				"type": label.Type,
			})
		}
		result = append(result, map[string]interface{}{
			"id":        item.ID,
			"name":      item.Name,
			"status":    item.Status,
			"busy":      item.Busy,
			"disabled":  item.Disabled,
			"ephemeral": item.Ephemeral,
			"labels":    labels,
		})
	}
	return result
}

func flattenActionRuns(items []*gitea.ActionWorkflowRun) []interface{} {
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"id":            item.ID,
			"display_title": item.DisplayTitle,
			"event":         item.Event,
			"head_branch":   item.HeadBranch,
			"head_sha":      item.HeadSha,
			"path":          item.Path,
			"run_attempt":   item.RunAttempt,
			"run_number":    item.RunNumber,
			"status":        item.Status,
			"conclusion":    item.Conclusion,
			"url":           item.URL,
			"html_url":      item.HTMLURL,
			"started_at":    item.StartedAt.Format(time.RFC3339),
			"completed_at":  timeToString(item.CompletedAt),
		})
	}
	return result
}

func flattenActionJobs(items []*gitea.ActionWorkflowJob) []interface{} {
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		steps := make([]interface{}, 0, len(item.Steps))
		for _, step := range item.Steps {
			if step == nil {
				continue
			}
			steps = append(steps, map[string]interface{}{
				"name":         step.Name,
				"number":       step.Number,
				"status":       step.Status,
				"conclusion":   step.Conclusion,
				"started_at":   timeToString(step.StartedAt),
				"completed_at": timeToString(step.CompletedAt),
			})
		}
		result = append(result, map[string]interface{}{
			"id":           item.ID,
			"run_id":       item.RunID,
			"run_url":      item.RunURL,
			"run_attempt":  item.RunAttempt,
			"name":         item.Name,
			"head_branch":  item.HeadBranch,
			"head_sha":     item.HeadSha,
			"status":       item.Status,
			"conclusion":   item.Conclusion,
			"url":          item.URL,
			"html_url":     item.HTMLURL,
			"created_at":   timeToString(item.CreatedAt),
			"started_at":   timeToString(item.StartedAt),
			"completed_at": timeToString(item.CompletedAt),
			"runner_id":    item.RunnerID,
			"runner_name":  item.RunnerName,
			"labels":       item.Labels,
			"steps":        steps,
		})
	}
	return result
}

func flattenActionWorkflows(items []*gitea.ActionWorkflow) []interface{} {
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"id":         item.ID,
			"name":       item.Name,
			"path":       item.Path,
			"state":      item.State,
			"created_at": timeToString(item.CreatedAt),
			"updated_at": timeToString(item.UpdatedAt),
			"url":        item.URL,
			"html_url":   item.HTMLURL,
			"badge_url":  item.BadgeURL,
			"deleted_at": timeToString(item.DeletedAt),
		})
	}
	return result
}

func flattenActionArtifacts(items []*gitea.ActionArtifact) []interface{} {
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		runID := int64(0)
		if item.WorkflowRun != nil {
			runID = item.WorkflowRun.ID
		}
		result = append(result, map[string]interface{}{
			"id":                   item.ID,
			"name":                 item.Name,
			"size_in_bytes":        item.SizeInBytes,
			"url":                  item.URL,
			"archive_download_url": item.ArchiveDownloadURL,
			"expired":              item.Expired,
			"workflow_run_id":      runID,
			"created_at":           timeToString(item.CreatedAt),
			"updated_at":           timeToString(item.UpdatedAt),
			"expires_at":           timeToString(item.ExpiresAt),
		})
	}
	return result
}

func flattenPackageVersions(items []*gitea.Package) []interface{} {
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		owner := ""
		if item.Owner != nil {
			owner = item.Owner.UserName
		}
		creator := ""
		if item.Creator != nil {
			creator = item.Creator.UserName
		}
		repository := ""
		if item.Repository != nil {
			repository = item.Repository.FullName
		}
		result = append(result, map[string]interface{}{
			"id":         item.ID,
			"owner":      owner,
			"repository": repository,
			"creator":    creator,
			"type":       item.Type,
			"name":       item.Name,
			"version":    item.Version,
			"html_url":   item.HTMLURL,
			"created_at": timeToString(item.CreatedAt),
		})
	}
	return result
}

func flattenContents(items []*gitea.ContentsResponse) []interface{} {
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"name":                item.Name,
			"path":                item.Path,
			"sha":                 item.SHA,
			"type":                item.Type,
			"size":                item.Size,
			"last_commit_sha":     derefString(item.LastCommitSha),
			"last_commit_message": derefString(item.LastCommitMessage),
			"last_committer_date": timePtrToString(item.LastCommitterDate),
			"last_author_date":    timePtrToString(item.LastAuthorDate),
			"encoding":            derefString(item.Encoding),
			"content":             derefString(item.Content),
			"target":              derefString(item.Target),
			"url":                 derefString(item.URL),
			"html_url":            derefString(item.HTMLURL),
			"git_url":             derefString(item.GitURL),
			"download_url":        derefString(item.DownloadURL),
			"submodule_git_url":   derefString(item.SubmoduleGitURL),
			"lfs_oid":             derefString(item.LfsOid),
			"lfs_size":            derefInt64(item.LfsSize),
		})
	}
	return result
}

func flattenPullRequest(pr *gitea.PullRequest) map[string]interface{} {
	if pr == nil {
		return map[string]interface{}{}
	}
	baseRef := ""
	headRef := ""
	if pr.Base != nil {
		baseRef = pr.Base.Ref
	}
	if pr.Head != nil {
		headRef = pr.Head.Ref
	}
	return map[string]interface{}{
		"id":           pr.ID,
		"index":        pr.Index,
		"title":        pr.Title,
		"body":         pr.Body,
		"state":        string(pr.State),
		"draft":        pr.Draft,
		"is_locked":    pr.IsLocked,
		"comments":     pr.Comments,
		"html_url":     pr.HTMLURL,
		"diff_url":     pr.DiffURL,
		"patch_url":    pr.PatchURL,
		"mergeable":    pr.Mergeable,
		"has_merged":   pr.HasMerged,
		"base_ref":     baseRef,
		"head_ref":     headRef,
		"merge_base":   pr.MergeBase,
		"created_at":   timePtrToString(pr.Created),
		"updated_at":   timePtrToString(pr.Updated),
		"closed_at":    timePtrToString(pr.Closed),
		"merged_at":    timePtrToString(pr.Merged),
		"merge_commit": derefString(pr.MergedCommitID),
	}
}

func copyAndSortStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefInt64(value *int64) int {
	if value == nil {
		return 0
	}
	return int(*value)
}

func timePtrToString(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(time.RFC3339)
}

func timeToString(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}

const maxPaginatedPages = 100

func collectPaginated[T any](fetch func(page int) ([]T, error)) ([]T, error) {
	return collectPaginatedWithLimit(maxPaginatedPages, fetch)
}

func collectPaginatedWithLimit[T any](maxPages int, fetch func(page int) ([]T, error)) ([]T, error) {
	if maxPages < 1 {
		return nil, fmt.Errorf("maxPages must be at least 1")
	}

	all := make([]T, 0)
	for page := 1; page <= maxPages+1; page++ {
		items, err := fetch(page)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return all, nil
		}
		if page > maxPages {
			return nil, fmt.Errorf("pagination exceeded safety limit of %d pages", maxPages)
		}
		all = append(all, items...)
	}

	return all, nil
}
