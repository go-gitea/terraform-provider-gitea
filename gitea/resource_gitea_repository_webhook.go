package gitea

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	repoWebhookUsername            string = "username"
	repoWebhookName                string = "name"
	repoWebhookType                string = "type"
	repoWebhookUrl                 string = "url"
	repoWebhookContentType         string = "content_type"
	repoWebhookSecret              string = "secret"
	repoWebhookAuthorizationHeader string = "authorization_header"
	repoWebhookEvents              string = "events"
	repoWebhookBranchFilter        string = "branch_filter"
	repoWebhookActive              string = "active"
	repoWebhookCreatedAt           string = "created_at"
	repoWebhookChannel             string = "channel"
	repoWebhookSlackUsername       string = "slack_username"
	repoWebhookIconUrl             string = "icon_url"
	repoWebhookColor               string = "color"
	repoWebhookHttpMethod          string = "http_method"
	repoWebhookConfig              string = "config"
)

func resourceRepositoryWebhookRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	user := d.Get(repoWebhookUsername).(string)
	repo := d.Get(repoWebhookName).(string)

	hook, resp, err := client.GetRepoHook(user, repo, id)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return
		} else {
			return err
		}
	}

	err = setRepositoryWebhookData(hook, d)

	return
}

func buildWebhookConfigMap(d *schema.ResourceData) map[string]string {
	config := make(map[string]string)

	if rawConfig, ok := d.GetOk(repoWebhookConfig); ok {
		for k, v := range rawConfig.(map[string]interface{}) {
			if strVal, ok := v.(string); ok {
				config[k] = strVal
			}
		}
	}

	if v, ok := d.GetOk(repoWebhookUrl); ok && v.(string) != "" {
		config["url"] = v.(string)
	}
	if v, ok := d.GetOk(repoWebhookContentType); ok && v.(string) != "" {
		config["content_type"] = v.(string)
	} else if wType, ok := d.GetOk(repoWebhookType); ok {
		t := strings.ToLower(wType.(string))
		if (t == "gitea" || t == "gogs") && config["content_type"] == "" {
			config["content_type"] = "json"
		}
	}
	if v, ok := d.GetOk(repoWebhookSecret); ok && v.(string) != "" {
		config["secret"] = v.(string)
	}
	if v, ok := d.GetOk(repoWebhookHttpMethod); ok && v.(string) != "" {
		config["http_method"] = v.(string)
	}
	if v, ok := d.GetOk(repoWebhookChannel); ok && v.(string) != "" {
		config["channel"] = v.(string)
	}
	if v, ok := d.GetOk(repoWebhookSlackUsername); ok && v.(string) != "" {
		config["username"] = v.(string)
	}
	if v, ok := d.GetOk(repoWebhookIconUrl); ok && v.(string) != "" {
		config["icon_url"] = v.(string)
	}
	if v, ok := d.GetOk(repoWebhookColor); ok && v.(string) != "" {
		config["color"] = v.(string)
	}

	return config
}

func resourceRepositoryWebhookCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	user := d.Get(repoWebhookUsername).(string)
	repo := d.Get(repoWebhookName).(string)

	config := buildWebhookConfigMap(d)

	events := make([]string, 0)
	for _, element := range d.Get(repoWebhookEvents).([]interface{}) {
		events = append(events, element.(string))
	}

	hookOption := gitea.CreateHookOption{
		Type:                gitea.HookType(d.Get(repoWebhookType).(string)),
		Config:              config,
		Events:              events,
		BranchFilter:        d.Get(repoWebhookBranchFilter).(string),
		Active:              d.Get(repoWebhookActive).(bool),
		AuthorizationHeader: d.Get(repoWebhookAuthorizationHeader).(string),
	}

	hook, _, err := client.CreateRepoHook(user, repo, hookOption)
	if err != nil {
		return err
	}

	err = setRepositoryWebhookData(hook, d)

	return
}

func resourceRepositoryWebhookUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	user := d.Get(repoWebhookUsername).(string)
	repo := d.Get(repoWebhookName).(string)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	config := buildWebhookConfigMap(d)

	events := make([]string, 0)
	for _, element := range d.Get(repoWebhookEvents).([]interface{}) {
		events = append(events, element.(string))
	}

	active := d.Get(repoWebhookActive).(bool)

	hookOption := gitea.EditHookOption{
		Config:              config,
		Events:              events,
		BranchFilter:        d.Get(repoWebhookBranchFilter).(string),
		Active:              &active,
		AuthorizationHeader: d.Get(repoWebhookAuthorizationHeader).(string),
	}

	_, err = client.EditRepoHook(user, repo, id, hookOption)
	if err != nil {
		return err
	}

	hook, _, err := client.GetRepoHook(user, repo, id)
	if err != nil {
		return err
	}

	err = setRepositoryWebhookData(hook, d)

	return
}

func resourceRepositoryWebhookDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	user := d.Get(repoWebhookUsername).(string)
	repo := d.Get(repoWebhookName).(string)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	_, err = client.DeleteRepoHook(user, repo, id)
	if err != nil {
		return err
	}

	return
}

func setRepositoryWebhookData(hook *gitea.Hook, d *schema.ResourceData) (err error) {
	d.SetId(strconv.FormatInt(hook.ID, 10))

	d.Set(repoWebhookUsername, d.Get(repoWebhookUsername).(string))
	d.Set(repoWebhookName, d.Get(repoWebhookName).(string))
	d.Set(repoWebhookType, hook.Type)
	d.Set(repoWebhookUrl, hookConfigValue(hook, "url"))
	d.Set(repoWebhookContentType, hookConfigValue(hook, "content_type"))

	secret := hookConfigValue(hook, "secret")
	if secret == "" {
		secret = d.Get(repoWebhookSecret).(string)
	}
	if secret != "" {
		d.Set(repoWebhookSecret, secret)
	}

	d.Set(repoWebhookEvents, stringSliceToInterfaceSlice(hook.Events))
	d.Set(repoWebhookBranchFilter, hook.BranchFilter)
	d.Set(repoWebhookActive, hook.Active)
	d.Set(repoWebhookCreatedAt, hook.Created)
	d.Set(repoWebhookAuthorizationHeader, hook.AuthorizationHeader)

	if v := hookConfigValue(hook, "http_method"); v != "" {
		d.Set(repoWebhookHttpMethod, v)
	}
	if v := hookConfigValue(hook, "channel"); v != "" {
		d.Set(repoWebhookChannel, v)
	}
	if v := hookConfigValue(hook, "username"); v != "" {
		d.Set(repoWebhookSlackUsername, v)
	}
	if v := hookConfigValue(hook, "icon_url"); v != "" {
		d.Set(repoWebhookIconUrl, v)
	}
	if v := hookConfigValue(hook, "color"); v != "" {
		d.Set(repoWebhookColor, v)
	}

	if hook.Config != nil {
		d.Set(repoWebhookConfig, hook.Config)
	}

	return
}

func hookConfigValue(hook *gitea.Hook, key string) string {
	if hook == nil || hook.Config == nil {
		return ""
	}
	return hook.Config[key]
}

func stringSliceToInterfaceSlice(values []string) []interface{} {
	result := make([]interface{}, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

func resourceGiteaRepositoryWebhook() *schema.Resource {
	return &schema.Resource{
		Read:   resourceRepositoryWebhookRead,
		Create: resourceRepositoryWebhookCreate,
		Update: resourceRepositoryWebhookUpdate,
		Delete: resourceRepositoryWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 3 {
					return nil, fmt.Errorf("unexpected ID format (%q), expected <username>/<repo>/<webhook_id>", d.Id())
				}
				d.Set("username", parts[0])
				d.Set("name", parts[1])
				d.SetId(parts[2])
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "User name or organization name",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Repository name",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Webhook type, e.g. `gitea`, `gogs`, `slack`, `discord`, `dingtalk`, `msteams`, `telegram`, `feishu`, `matrix`, `wechatwork`, `packagist`",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Target URL of the webhook",
			},
			"content_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The content type of the payload. It can be `json`, or `form`",
			},
			"secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Webhook secret",
			},
			"authorization_header": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Webhook authorization header",
			},
			"events": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "A list of events that will trigger the webhook, e.g. `[\"push\"]`",
			},
			"branch_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "*",
				Description: "Set branch filter on the webhook, e.g. `\"*\"`",
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Set webhook to active, e.g. `true`",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Webhook creation timestamp",
			},
			"channel": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Channel name for Slack webhooks (e.g. `#general` or `@username`)",
			},
			"slack_username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Bot username for Slack webhooks",
			},
			"icon_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Icon URL for Slack or Discord webhooks",
			},
			"color": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Hex color code for Slack webhooks (e.g. `#ff0000`)",
			},
			"http_method": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "HTTP method used for the webhook",
			},
			"config": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Additional key-value configuration options for webhooks",
			},
		},
		Description: "This resource allows you to create and manage webhooks for repositories.",
	}
}
