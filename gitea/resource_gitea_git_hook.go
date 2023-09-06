package gitea

import (
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	GitHookUser    string = "user"
	GitHookRepo    string = "repo"
	GitHookName    string = "name"
	GitHookContent string = "content"
)

func resourceGitHookRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	user := d.Get(GitHookUser).(string)
	repo := d.Get(GitHookRepo).(string)
	name := d.Get(GitHookName).(string)

	gitHook, _, err := client.GetRepoGitHook(user, repo, name)

	if err != nil {
		return err
	}

	err = setGitHookResourceData(user, repo, gitHook, d)

	return
}

func resourceGitHookUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	user := d.Get(GitHookUser).(string)
	repo := d.Get(GitHookRepo).(string)
	name := d.Get(GitHookName).(string)

	opts := gitea.EditGitHookOption{
		Content: d.Get(GitHookContent).(string),
	}

	_, err = client.EditRepoGitHook(user, repo, name, opts)

	if err != nil {
		return err
	}

	// Get gitHook ourselves, EditRepoGitHook does not return it
	gitHook, _, err := client.GetRepoGitHook(user, repo, name)

	if err != nil {
		return err
	}

	err = setGitHookResourceData(user, repo, gitHook, d)

	return
}

func resourceGitHookDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	user := d.Get(GitHookUser).(string)
	repo := d.Get(GitHookRepo).(string)
	name := d.Get(GitHookName).(string)

	_, err = client.DeleteRepoGitHook(user, repo, name)

	return
}

func setGitHookResourceData(user string, repo string, gitHook *gitea.GitHook, d *schema.ResourceData) (err error) {
	d.SetId(fmt.Sprintf("%s/%s/%s", user, repo, gitHook.Name))
	d.Set(GitHookUser, user)
	d.Set(GitHookRepo, repo)
	d.Set(GitHookName, gitHook.Name)
	d.Set(GitHookContent, gitHook.Content)
	return
}

func resourceGiteaGitHook() *schema.Resource {
	return &schema.Resource{
		Read:   resourceGitHookRead,
		Create: resourceGitHookUpdate, // All hooks already exist, just empty and disabled
		Update: resourceGitHookUpdate,
		Delete: resourceGitHookDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the git hook to configure",
			},
			"repo": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The repository that this hook belongs too.",
			},
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The user (or organisation) owning the repo this hook belongs too",
			},
			"content": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Content of the git hook",
			},
		},
		Description: "`gitea_git_hook` manages git hooks on a repository.\n" +
			"import is currently not supported\n\n" +
			"WARNING: using this resource requires to enable server side hooks" +
			"which are known to cause [security issues](https://github.com/go-gitea/gitea/pull/13058)!\n\n" +
			"if you want to procede, you need to enable server side hooks as stated" +
			" [here](https://docs.gitea.io/en-us/config-cheat-sheet/#security-security)",
	}
}
