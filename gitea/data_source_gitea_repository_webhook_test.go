package gitea

import (
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceGiteaRepositoryWebhookRead(t *testing.T) {
	d := schema.TestResourceDataRaw(t, dataSourceGiteaRepositoryWebhook().Schema, map[string]interface{}{
		"username": "owner-name",
		"name":     "repo-name",
		"id":       100,
	})

	hook := &gitea.Hook{
		ID:     100,
		Type:   "slack",
		Events: []string{"push", "issues"},
		Config: map[string]string{
			"url":      "https://hooks.slack.com/services/test",
			"channel":  "#alerts",
			"username": "gitea-notifier",
			"icon_url": "https://example.com/logo.png",
			"color":    "#36a64f",
		},
		BranchFilter:        "main",
		Active:              true,
		Created:             time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		AuthorizationHeader: "Bearer test",
	}

	d.SetId("100")
	d.Set("username", "owner-name")
	d.Set("name", "repo-name")
	d.Set("type", hook.Type)
	d.Set("url", hookConfigValue(hook, "url"))
	d.Set("content_type", hookConfigValue(hook, "content_type"))
	d.Set("secret", hookConfigValue(hook, "secret"))
	d.Set("authorization_header", hook.AuthorizationHeader)
	d.Set("events", stringSliceToInterfaceSlice(hook.Events))
	d.Set("branch_filter", hook.BranchFilter)
	d.Set("active", hook.Active)
	d.Set("created_at", hook.Created.Format("2006-01-02T15:04:05Z07:00"))

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

	if got := d.Get("channel").(string); got != "#alerts" {
		t.Fatalf("expected channel #alerts, got %q", got)
	}
	if got := d.Get("slack_username").(string); got != "gitea-notifier" {
		t.Fatalf("expected slack_username gitea-notifier, got %q", got)
	}
	if got := d.Get("color").(string); got != "#36a64f" {
		t.Fatalf("expected color #36a64f, got %q", got)
	}
	if got := d.Get("type").(string); got != "slack" {
		t.Fatalf("expected type slack, got %q", got)
	}
}
