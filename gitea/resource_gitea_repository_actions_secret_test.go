package gitea

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRepositoryActionsSecretReadReturnsServerErrorWithoutClearingState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/owner/repo/actions/secrets" {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := gitea.NewClient(server.URL, gitea.SetGiteaVersion("1.25.0"))
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}

	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryActionsSecret().Schema, map[string]interface{}{})
	d.SetId(buildThreePartID("owner", "repo", "secret"))

	err = resourceGiteaRepositoryActionsSecretRead(d, client)
	if err == nil {
		t.Fatal("expected read to return an error")
	}
	if d.Id() == "" {
		t.Fatal("expected state ID to be preserved on server error")
	}
}
