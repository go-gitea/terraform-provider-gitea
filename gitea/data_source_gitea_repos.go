package gitea

import (
	"crypto/sha256"
	"fmt"
	"log"
	"sort"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaRepos() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaReposRead,
		Schema: map[string]*schema.Schema{
			"owner": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Restrict results to repositories owned by this user or organization.",
			},
			"query": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Free-text keyword to match against repository name (and description when keyword_in_description is set).",
			},
			"keyword_is_topic": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Treat 'query' as a topic name rather than a free-text keyword.",
			},
			"keyword_in_description": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Also match 'query' against repository descriptions, not just names.",
			},
			"private": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "all",
				Description: "Filter by visibility: 'true' (private only), 'false' (public only), or 'all' (default).",
				ValidateFunc: func(v interface{}, key string) (ws []string, es []error) {
					// [LAW:dataflow-not-control-flow] enumerated values map to a tri-state filter passed to the SDK.
					switch v.(string) {
					case "all", "true", "false":
					default:
						es = append(es, fmt.Errorf("%s must be one of 'all', 'true', 'false'", key))
					}
					return
				},
			},
			"archived": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "all",
				Description: "Filter by archived state: 'true' (archived only), 'false' (active only), or 'all' (default).",
				ValidateFunc: func(v interface{}, key string) (ws []string, es []error) {
					switch v.(string) {
					case "all", "true", "false":
					default:
						es = append(es, fmt.Errorf("%s must be one of 'all', 'true', 'false'", key))
					}
					return
				},
			},
			"exclude_template": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "When true, template repositories are excluded from results.",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "",
				Description: "Restrict results to a repo type: '' (any), 'source', 'fork', or 'mirror'.",
				ValidateFunc: func(v interface{}, key string) (ws []string, es []error) {
					switch v.(string) {
					case "", "source", "fork", "mirror":
					default:
						es = append(es, fmt.Errorf("%s must be one of '', 'source', 'fork', 'mirror'", key))
					}
					return
				},
			},
			"page_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     50,
				ForceNew:    true,
				Description: "Page size for the underlying SearchRepos API. Pagination is handled internally.",
			},
			"repos": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":             {Type: schema.TypeInt, Computed: true},
						"name":           {Type: schema.TypeString, Computed: true},
						"owner":          {Type: schema.TypeString, Computed: true},
						"full_name":      {Type: schema.TypeString, Computed: true},
						"description":    {Type: schema.TypeString, Computed: true},
						"private":        {Type: schema.TypeBool, Computed: true},
						"fork":           {Type: schema.TypeBool, Computed: true},
						"mirror":         {Type: schema.TypeBool, Computed: true},
						"template":       {Type: schema.TypeBool, Computed: true},
						"archived":       {Type: schema.TypeBool, Computed: true},
						"empty":          {Type: schema.TypeBool, Computed: true},
						"default_branch": {Type: schema.TypeString, Computed: true},
						"html_url":       {Type: schema.TypeString, Computed: true},
						"ssh_url":        {Type: schema.TypeString, Computed: true},
						"clone_url":      {Type: schema.TypeString, Computed: true},
						"website":        {Type: schema.TypeString, Computed: true},
					},
				},
			},
			"full_names": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Convenience flat list of full_name values, in the same order as 'repos'.",
			},
		},
	}
}

func dataSourceGiteaReposRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	opts, err := buildSearchRepoOptions(client, d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Listing Gitea repos with filters: owner=%q query=%q private=%q archived=%q type=%q",
		d.Get("owner").(string),
		d.Get("query").(string),
		d.Get("private").(string),
		d.Get("archived").(string),
		d.Get("type").(string),
	)

	all, err := searchAllRepos(client, opts)
	if err != nil {
		return err
	}

	repos := make([]map[string]interface{}, 0, len(all))
	fullNames := make([]string, 0, len(all))
	for _, r := range all {
		repos = append(repos, flattenRepo(r))
		fullNames = append(fullNames, r.FullName)
	}

	if err := d.Set("repos", repos); err != nil {
		return err
	}
	if err := d.Set("full_names", fullNames); err != nil {
		return err
	}

	d.SetId(filterDigest(d))
	return nil
}

// buildSearchRepoOptions translates the data-source schema values into a
// SearchRepoOptions in one place. Resolving an owner name to a uid happens
// here because the SDK's filter is numeric while the user-facing API is
// name-based — keeping the translation at the boundary.
// [LAW:single-enforcer] only one place builds the SDK options.
func buildSearchRepoOptions(client *gitea.Client, d *schema.ResourceData) (gitea.SearchRepoOptions, error) {
	opts := gitea.SearchRepoOptions{
		Keyword:              d.Get("query").(string),
		KeywordIsTopic:       d.Get("keyword_is_topic").(bool),
		KeywordInDescription: d.Get("keyword_in_description").(bool),
		ExcludeTemplate:      d.Get("exclude_template").(bool),
		Type:                 gitea.RepoType(d.Get("type").(string)),
		ListOptions:          gitea.ListOptions{PageSize: d.Get("page_size").(int)},
	}

	opts.IsPrivate = triStateToBoolPtr(d.Get("private").(string))
	opts.IsArchived = triStateToBoolPtr(d.Get("archived").(string))

	if owner := strings.TrimSpace(d.Get("owner").(string)); owner != "" {
		uid, err := resolveOwnerID(client, owner)
		if err != nil {
			return opts, fmt.Errorf("resolving owner %q: %w", owner, err)
		}
		opts.OwnerID = uid
	}

	return opts, nil
}

// triStateToBoolPtr maps the "all"/"true"/"false" schema values to the SDK's
// *bool optional. "all" → nil (don't filter); "true"/"false" → corresponding bool.
func triStateToBoolPtr(v string) *bool {
	switch strings.ToLower(v) {
	case "true":
		t := true
		return &t
	case "false":
		f := false
		return &f
	default:
		return nil
	}
}

// resolveOwnerID looks up a user or organization by name and returns its uid.
// Gitea models orgs as users in the same id space, so a single lookup is
// authoritative for both. [LAW:one-source-of-truth] no parallel user/org paths.
func resolveOwnerID(client *gitea.Client, name string) (int64, error) {
	user, _, err := client.GetUserInfo(name)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

// searchAllRepos walks every page of SearchRepos and returns the concatenated
// result. The same operations execute every iteration; the page number is the
// only datum that varies. [LAW:dataflow-not-control-flow]
func searchAllRepos(client *gitea.Client, opts gitea.SearchRepoOptions) ([]*gitea.Repository, error) {
	var all []*gitea.Repository
	page := 1
	for {
		opts.Page = page
		repos, resp, err := client.SearchRepos(opts)
		if err != nil {
			return nil, err
		}
		all = append(all, repos...)
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return all, nil
}

func flattenRepo(r *gitea.Repository) map[string]interface{} {
	owner := ""
	if r.Owner != nil {
		owner = r.Owner.UserName
	}
	return map[string]interface{}{
		"id":             int(r.ID),
		"name":           r.Name,
		"owner":          owner,
		"full_name":      r.FullName,
		"description":    r.Description,
		"private":        r.Private,
		"fork":           r.Fork,
		"mirror":         r.Mirror,
		"template":       r.Template,
		"archived":       r.Archived,
		"empty":          r.Empty,
		"default_branch": r.DefaultBranch,
		"html_url":       r.HTMLURL,
		"ssh_url":        r.SSHURL,
		"clone_url":      r.CloneURL,
		"website":        r.Website,
	}
}

// filterDigest produces a stable id from the filter inputs so the same query
// yields the same id across refreshes. Listing the keys explicitly keeps the
// digest tied to the documented filter surface — adding a new filter is the
// one place that needs an update. [LAW:one-source-of-truth]
func filterDigest(d *schema.ResourceData) string {
	keys := []string{
		"owner", "query", "keyword_is_topic", "keyword_in_description",
		"private", "archived", "exclude_template", "type", "page_size",
	}
	sort.Strings(keys)
	h := sha256.New()
	for _, k := range keys {
		fmt.Fprintf(h, "%s=%v\n", k, d.Get(k))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
