package gitea

import (
	"fmt"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaRepos() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaReposRead,
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Owner whose repositories to list.",
			},
			"repositories": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":             {Type: schema.TypeInt, Computed: true},
						"name":           {Type: schema.TypeString, Computed: true},
						"full_name":      {Type: schema.TypeString, Computed: true},
						"description":    {Type: schema.TypeString, Computed: true},
						"private":        {Type: schema.TypeBool, Computed: true},
						"fork":           {Type: schema.TypeBool, Computed: true},
						"mirror":         {Type: schema.TypeBool, Computed: true},
						"size":           {Type: schema.TypeInt, Computed: true},
						"html_url":       {Type: schema.TypeString, Computed: true},
						"ssh_url":        {Type: schema.TypeString, Computed: true},
						"clone_url":      {Type: schema.TypeString, Computed: true},
						"website":        {Type: schema.TypeString, Computed: true},
						"stars":          {Type: schema.TypeInt, Computed: true},
						"forks":          {Type: schema.TypeInt, Computed: true},
						"watchers":       {Type: schema.TypeInt, Computed: true},
						"open_issues":    {Type: schema.TypeInt, Computed: true},
						"default_branch": {Type: schema.TypeString, Computed: true},
						"archived":       {Type: schema.TypeBool, Computed: true},
						"created":        {Type: schema.TypeString, Computed: true},
						"updated":        {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceGiteaReposRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	username := d.Get("username").(string)

	repos, err := collectPaginated(func(page int) ([]*gitea.Repository, error) {
		items, _, err := client.ListUserRepos(username, gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
		})
		return items, err
	})
	if err != nil {
		return fmt.Errorf("unable to list repos for user %q: %w", username, err)
	}

	result := make([]interface{}, 0, len(repos))
	for _, repo := range repos {
		result = append(result, map[string]interface{}{
			"id":             int(repo.ID),
			"name":           repo.Name,
			"full_name":      repo.FullName,
			"description":    repo.Description,
			"private":        repo.Private,
			"fork":           repo.Fork,
			"mirror":         repo.Mirror,
			"size":           repo.Size,
			"html_url":       repo.HTMLURL,
			"ssh_url":        repo.SSHURL,
			"clone_url":      repo.CloneURL,
			"website":        repo.Website,
			"stars":          repo.Stars,
			"forks":          repo.Forks,
			"watchers":       repo.Watchers,
			"open_issues":    repo.OpenIssues,
			"default_branch": repo.DefaultBranch,
			"archived":       repo.Archived,
			"created":        repo.Created.Format(time.RFC3339),
			"updated":        repo.Updated.Format(time.RFC3339),
		})
	}

	d.SetId(fmt.Sprintf("user:%s:repos", username))
	if err := d.Set("repositories", result); err != nil {
		return fmt.Errorf("failed to set repositories: %w", err)
	}

	return nil
}
