package gitea

import (
	"fmt"
	"net/http"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaRepositoryFiles() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaRepositoryFilesRead,
		Schema: mergeSchemaMaps(repositoryIdentitySchema(), map[string]*schema.Schema{
			"branch": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The branch or ref to read from.",
			},
			"paths": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "The repository paths to fetch.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"files": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The fetched file metadata and contents.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name":                {Type: schema.TypeString, Computed: true},
						"path":                {Type: schema.TypeString, Computed: true},
						"sha":                 {Type: schema.TypeString, Computed: true},
						"type":                {Type: schema.TypeString, Computed: true},
						"size":                {Type: schema.TypeInt, Computed: true},
						"last_commit_sha":     {Type: schema.TypeString, Computed: true},
						"last_commit_message": {Type: schema.TypeString, Computed: true},
						"last_committer_date": {Type: schema.TypeString, Computed: true},
						"last_author_date":    {Type: schema.TypeString, Computed: true},
						"encoding":            {Type: schema.TypeString, Computed: true},
						"content":             {Type: schema.TypeString, Computed: true},
						"target":              {Type: schema.TypeString, Computed: true},
						"url":                 {Type: schema.TypeString, Computed: true},
						"html_url":            {Type: schema.TypeString, Computed: true},
						"git_url":             {Type: schema.TypeString, Computed: true},
						"download_url":        {Type: schema.TypeString, Computed: true},
						"submodule_git_url":   {Type: schema.TypeString, Computed: true},
						"lfs_oid":             {Type: schema.TypeString, Computed: true},
						"lfs_size":            {Type: schema.TypeInt, Computed: true},
					},
				},
			},
		}),
		Description: "Fetches multiple repository files in a single request.",
	}
}

func dataSourceGiteaRepositoryFilesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repo := strings.ToLower(d.Get(repositoryNameField).(string))
	branch := d.Get("branch").(string)

	rawPaths := d.Get("paths").([]interface{})
	paths := make([]string, 0, len(rawPaths))
	for _, rawPath := range rawPaths {
		paths = append(paths, strings.TrimPrefix(rawPath.(string), "/"))
	}

	files, _, err := client.PostRepoFileContents(owner, repo, branch, gitea.GetFilesOptions{Files: paths})
	if err != nil {
		return err
	}

	if err := d.Set("files", flattenContents(files)); err != nil {
		return fmt.Errorf("error setting files: %w", err)
	}
	d.SetId(buildResourceID(owner, repo, branch, strings.Join(copyAndSortStrings(paths), "|")))
	return nil
}

func dataSourceGiteaPackageVersions() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaPackageVersionsRead,
		Schema: map[string]*schema.Schema{
			"owner": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The package owner.",
			},
			"package_type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The package type.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The package name.",
			},
			"versions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "All discovered versions for the package.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":         {Type: schema.TypeInt, Computed: true},
						"owner":      {Type: schema.TypeString, Computed: true},
						"repository": {Type: schema.TypeString, Computed: true},
						"creator":    {Type: schema.TypeString, Computed: true},
						"type":       {Type: schema.TypeString, Computed: true},
						"name":       {Type: schema.TypeString, Computed: true},
						"version":    {Type: schema.TypeString, Computed: true},
						"html_url":   {Type: schema.TypeString, Computed: true},
						"created_at": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
		Description: "Lists all versions of a package.",
	}
}

func dataSourceGiteaPackageVersionsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	owner := d.Get("owner").(string)
	packageType := d.Get("package_type").(string)
	name := d.Get("name").(string)

	packages, err := collectPaginated(func(page int) ([]*gitea.Package, error) {
		items, _, callErr := client.ListPackageVersions(owner, packageType, name, gitea.ListPackagesOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 100},
		})
		return items, callErr
	})
	if err != nil {
		return err
	}

	if err := d.Set("versions", flattenPackageVersions(packages)); err != nil {
		return fmt.Errorf("error setting versions: %w", err)
	}
	d.SetId(buildResourceID(owner, packageType, name))
	return nil
}

func dataSourceGiteaRepositoryIssueConfig() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaRepositoryIssueConfigRead,
		Schema: mergeSchemaMaps(repositoryIdentitySchema(), map[string]*schema.Schema{
			"blank_issues_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether blank issues are allowed.",
			},
			"contact_links": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Configured contact links.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name":  {Type: schema.TypeString, Computed: true},
						"url":   {Type: schema.TypeString, Computed: true},
						"about": {Type: schema.TypeString, Computed: true},
					},
				},
			},
			"valid": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the repository issue config validates.",
			},
			"validation_message": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Validation details returned by Gitea.",
			},
		}),
		Description: "Fetches the parsed issue configuration for a repository.",
	}
}

func dataSourceGiteaRepositoryIssueConfigRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repo := strings.ToLower(d.Get(repositoryNameField).(string))

	config, _, err := client.GetIssueConfig(owner, repo)
	if err != nil {
		return err
	}
	validation, _, err := client.ValidateIssueConfig(owner, repo)
	if err != nil {
		return err
	}

	d.Set("blank_issues_enabled", config.BlankIssuesEnabled)
	if err := d.Set("contact_links", flattenContactLinks(config.ContactLinks)); err != nil {
		return fmt.Errorf("error setting contact_links: %w", err)
	}
	d.Set("valid", validation.Valid)
	d.Set("validation_message", validation.Message)
	d.SetId(buildResourceID(owner, repo))
	return nil
}

func dataSourceGiteaRepositoryLicenses() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaRepositoryLicensesRead,
		Schema: mergeSchemaMaps(repositoryIdentitySchema(), map[string]*schema.Schema{
			"licenses": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Detected repository licenses.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		}),
		Description: "Lists detected licenses for a repository.",
	}
}

func dataSourceGiteaRepositoryLicensesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repo := strings.ToLower(d.Get(repositoryNameField).(string))
	licenses, _, err := client.GetRepoLicenses(owner, repo)
	if err != nil {
		return err
	}
	if err := d.Set("licenses", copyAndSortStrings(licenses)); err != nil {
		return fmt.Errorf("error setting licenses: %w", err)
	}
	d.SetId(buildResourceID(owner, repo))
	return nil
}

func dataSourceGiteaRepositorySigningKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaRepositorySigningKeyRead,
		Schema: mergeSchemaMaps(repositoryIdentitySchema(), map[string]*schema.Schema{
			"gpg_public_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The repository GPG signing key, if available.",
			},
			"ssh_public_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The repository SSH signing key, if available.",
			},
		}),
		Description: "Fetches repository signing keys.",
	}
}

func dataSourceGiteaRepositorySigningKeyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repo := strings.ToLower(d.Get(repositoryNameField).(string))

	gpgKey, gpgResp, gpgErr := client.GetRepoSigningKeyGPG(owner, repo)
	if gpgErr != nil && (gpgResp == nil || gpgResp.StatusCode != http.StatusNotFound) {
		return gpgErr
	}
	sshKey, sshResp, sshErr := client.GetRepoSigningKeySSH(owner, repo)
	if sshErr != nil && (sshResp == nil || sshResp.StatusCode != http.StatusNotFound) {
		return sshErr
	}

	d.Set("gpg_public_key", strings.TrimSpace(gpgKey))
	d.Set("ssh_public_key", strings.TrimSpace(sshKey))
	d.SetId(buildResourceID(owner, repo))
	return nil
}

func dataSourceGiteaRepositorySubscribers() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaRepositorySubscribersRead,
		Schema: mergeSchemaMaps(repositoryIdentitySchema(), map[string]*schema.Schema{
			"subscribers": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Repository subscribers.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":         {Type: schema.TypeInt, Computed: true},
						"username":   {Type: schema.TypeString, Computed: true},
						"full_name":  {Type: schema.TypeString, Computed: true},
						"email":      {Type: schema.TypeString, Computed: true},
						"avatar_url": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		}),
		Description: "Lists repository subscribers.",
	}
}

func dataSourceGiteaRepositorySubscribersRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repo := strings.ToLower(d.Get(repositoryNameField).(string))

	subscribers, err := collectPaginated(func(page int) ([]*gitea.User, error) {
		items, _, callErr := client.ListRepoSubscribers(owner, repo, gitea.ListOptions{Page: page, PageSize: 100})
		return items, callErr
	})
	if err != nil {
		return err
	}

	if err := d.Set("subscribers", flattenUsers(subscribers)); err != nil {
		return fmt.Errorf("error setting subscribers: %w", err)
	}
	d.SetId(buildResourceID(owner, repo))
	return nil
}

func dataSourceGiteaTeamRepository() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaTeamRepositoryRead,
		Schema: mergeSchemaMaps(repositoryIdentitySchema(), map[string]*schema.Schema{
			"team_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The team ID.",
			},
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"full_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"html_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_branch": {
				Type:     schema.TypeString,
				Computed: true,
			},
		}),
		Description: "Gets a repository attached to a team.",
	}
}

func dataSourceGiteaTeamRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	teamID := int64(d.Get("team_id").(int))
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repoName := strings.ToLower(d.Get(repositoryNameField).(string))
	repo, resp, err := client.GetTeamRepository(teamID, owner, repoName)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("repository %s/%s is not attached to team %d", owner, repoName, teamID)
		}
		return err
	}
	d.Set("id", int(repo.ID))
	d.Set("full_name", repo.FullName)
	d.Set("private", repo.Private)
	d.Set("html_url", repo.HTMLURL)
	d.Set("default_branch", repo.DefaultBranch)
	d.SetId(buildResourceID(fmt.Sprintf("%d", teamID), owner, repoName))
	return nil
}

func dataSourceGiteaPullRequestByBaseHead() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaPullRequestByBaseHeadRead,
		Schema: mergeSchemaMaps(repositoryIdentitySchema(), map[string]*schema.Schema{
			"base_ref": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The base branch name.",
			},
			"head_ref": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The head branch name.",
			},
			"pull_request": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The matching pull request.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":           {Type: schema.TypeInt, Computed: true},
						"index":        {Type: schema.TypeInt, Computed: true},
						"title":        {Type: schema.TypeString, Computed: true},
						"body":         {Type: schema.TypeString, Computed: true},
						"state":        {Type: schema.TypeString, Computed: true},
						"draft":        {Type: schema.TypeBool, Computed: true},
						"is_locked":    {Type: schema.TypeBool, Computed: true},
						"comments":     {Type: schema.TypeInt, Computed: true},
						"html_url":     {Type: schema.TypeString, Computed: true},
						"diff_url":     {Type: schema.TypeString, Computed: true},
						"patch_url":    {Type: schema.TypeString, Computed: true},
						"mergeable":    {Type: schema.TypeBool, Computed: true},
						"has_merged":   {Type: schema.TypeBool, Computed: true},
						"base_ref":     {Type: schema.TypeString, Computed: true},
						"head_ref":     {Type: schema.TypeString, Computed: true},
						"merge_base":   {Type: schema.TypeString, Computed: true},
						"created_at":   {Type: schema.TypeString, Computed: true},
						"updated_at":   {Type: schema.TypeString, Computed: true},
						"closed_at":    {Type: schema.TypeString, Computed: true},
						"merged_at":    {Type: schema.TypeString, Computed: true},
						"merge_commit": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		}),
		Description: "Fetches a pull request by base and head branch names.",
	}
}

func dataSourceGiteaPullRequestByBaseHeadRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repo := strings.ToLower(d.Get(repositoryNameField).(string))
	baseRef := d.Get("base_ref").(string)
	headRef := d.Get("head_ref").(string)

	pr, _, err := client.GetPullRequestByBaseHead(owner, repo, baseRef, headRef)
	if err != nil {
		return err
	}

	if err := d.Set("pull_request", []interface{}{flattenPullRequest(pr)}); err != nil {
		return fmt.Errorf("error setting pull_request: %w", err)
	}
	d.SetId(buildResourceID(owner, repo, baseRef, headRef))
	return nil
}
