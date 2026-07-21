package gitea

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaRepositoryWebhook() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaRepositoryWebhookRead,

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
			"id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Webhook ID",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Webhook type, e.g. `gitea`, `slack`, etc.",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Target URL of the webhook",
			},
			"content_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Payload content type (`json` or `form`)",
			},
			"secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Webhook secret",
			},
			"authorization_header": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Webhook authorization header",
			},
			"http_method": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "HTTP method used for the webhook",
			},
			"events": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of events that trigger the webhook",
			},
			"branch_filter": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Branch filter on the webhook",
			},
			"active": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the webhook is active",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Webhook creation timestamp",
			},
			"channel": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Channel name for Slack webhooks",
			},
			"slack_username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Bot username for Slack webhooks",
			},
			"icon_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Icon URL for Slack or Discord webhooks",
			},
			"color": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Hex color code for Slack webhooks",
			},
			"config": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Key-value configuration map for the webhook",
			},
		},
		Description: "Fetches details of a specific repository webhook.",
	}
}

func dataSourceGiteaRepositoryWebhookRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	owner := strings.ToLower(d.Get("username").(string))
	repo := strings.ToLower(d.Get("name").(string))
	id := int64(d.Get("id").(int))

	hook, resp, err := client.GetRepoHook(owner, repo, id)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("webhook with id %d not found for repo %s/%s", id, owner, repo)
		}
		return err
	}

	d.SetId(strconv.FormatInt(hook.ID, 10))
	d.Set("username", owner)
	d.Set("name", repo)
	d.Set("type", hook.Type)
	d.Set("url", hookConfigValue(hook, "url"))
	d.Set("content_type", hookConfigValue(hook, "content_type"))
	d.Set("secret", hookConfigValue(hook, "secret"))
	d.Set("authorization_header", hook.AuthorizationHeader)
	d.Set("events", stringSliceToInterfaceSlice(hook.Events))
	d.Set("branch_filter", hook.BranchFilter)
	d.Set("active", hook.Active)
	d.Set("created_at", hook.Created.Format("2006-01-02T15:04:05Z07:00"))

	if v := hookConfigValue(hook, "http_method"); v != "" {
		d.Set("http_method", v)
	}
	if v := hookConfigValue(hook, "channel"); v != "" {
		d.Set("channel", v)
	}
	if v := hookConfigValue(hook, "username"); v != "" {
		d.Set("slack_username", v)
	}
	if v := hookConfigValue(hook, "icon_url"); v != "" {
		d.Set("icon_url", v)
	}
	if v := hookConfigValue(hook, "color"); v != "" {
		d.Set("color", v)
	}
	if hook.Config != nil {
		d.Set("config", hook.Config)
	}

	return nil
}
