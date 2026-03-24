package gitea

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaActionsRunners() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaActionsRunnersRead,
		Schema: mergeSchemaMaps(actionScopeSchema(), map[string]*schema.Schema{
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Filter runners by disabled state.",
			},
			"runners": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The matching Actions runners.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":        {Type: schema.TypeInt, Computed: true},
						"name":      {Type: schema.TypeString, Computed: true},
						"status":    {Type: schema.TypeString, Computed: true},
						"busy":      {Type: schema.TypeBool, Computed: true},
						"disabled":  {Type: schema.TypeBool, Computed: true},
						"ephemeral": {Type: schema.TypeBool, Computed: true},
						"labels": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id":   {Type: schema.TypeInt, Computed: true},
									"name": {Type: schema.TypeString, Computed: true},
									"type": {Type: schema.TypeString, Computed: true},
								},
							},
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The total number of runners returned.",
			},
		}),
		Description: "Lists Actions runners for the selected scope.",
	}
}

func dataSourceGiteaActionsRunnersRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	cfg, err := resolveActionScope(d)
	if err != nil {
		return err
	}
	if err := requireVersion(client, ">= 1.25.0", "actions runners"); err != nil {
		return err
	}

	disabled := optionalBoolValue(d, "disabled")

	runners, err := collectPaginated(func(page int) ([]*gitea.ActionRunner, error) {
		response, callErr := listActionRunnersByScope(client, cfg, gitea.ListActionRunnersOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 100},
			Disabled:    disabled,
		})
		if callErr != nil {
			return nil, callErr
		}
		return response.Runners, nil
	})
	if err != nil {
		return err
	}

	if err := d.Set("runners", flattenActionRunners(runners)); err != nil {
		return fmt.Errorf("error setting runners: %w", err)
	}
	d.Set("total_count", len(runners))
	d.SetId(buildResourceID(cfg.Scope, cfg.Org, cfg.Owner, cfg.Repo))
	return nil
}

func dataSourceGiteaActionsRuns() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaActionsRunsRead,
		Schema: mergeSchemaMaps(actionScopeSchema(), map[string]*schema.Schema{
			"branch":   {Type: schema.TypeString, Optional: true},
			"event":    {Type: schema.TypeString, Optional: true},
			"status":   {Type: schema.TypeString, Optional: true},
			"actor":    {Type: schema.TypeString, Optional: true},
			"head_sha": {Type: schema.TypeString, Optional: true},
			"workflow_runs": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The matching workflow runs.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":            {Type: schema.TypeInt, Computed: true},
						"display_title": {Type: schema.TypeString, Computed: true},
						"event":         {Type: schema.TypeString, Computed: true},
						"head_branch":   {Type: schema.TypeString, Computed: true},
						"head_sha":      {Type: schema.TypeString, Computed: true},
						"path":          {Type: schema.TypeString, Computed: true},
						"run_attempt":   {Type: schema.TypeInt, Computed: true},
						"run_number":    {Type: schema.TypeInt, Computed: true},
						"status":        {Type: schema.TypeString, Computed: true},
						"conclusion":    {Type: schema.TypeString, Computed: true},
						"url":           {Type: schema.TypeString, Computed: true},
						"html_url":      {Type: schema.TypeString, Computed: true},
						"started_at":    {Type: schema.TypeString, Computed: true},
						"completed_at":  {Type: schema.TypeString, Computed: true},
					},
				},
			},
			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		}),
		Description: "Lists Actions workflow runs for the selected scope.",
	}
}

func dataSourceGiteaActionsRunsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	cfg, err := resolveActionScope(d)
	if err != nil {
		return err
	}
	if err := requireVersion(client, ">= 1.26.0", "actions runs"); err != nil {
		return err
	}

	opt := gitea.ListRepoActionRunsOptions{
		Branch:  d.Get("branch").(string),
		Event:   d.Get("event").(string),
		Status:  d.Get("status").(string),
		Actor:   d.Get("actor").(string),
		HeadSHA: d.Get("head_sha").(string),
	}

	runs, err := collectPaginated(func(page int) ([]*gitea.ActionWorkflowRun, error) {
		opt.ListOptions = gitea.ListOptions{Page: page, PageSize: 100}
		response, callErr := listActionRunsByScope(client, cfg, opt)
		if callErr != nil {
			return nil, callErr
		}
		return response.WorkflowRuns, nil
	})
	if err != nil {
		return err
	}

	if err := d.Set("workflow_runs", flattenActionRuns(runs)); err != nil {
		return fmt.Errorf("error setting workflow_runs: %w", err)
	}
	d.Set("total_count", len(runs))
	d.SetId(buildResourceID(cfg.Scope, cfg.Org, cfg.Owner, cfg.Repo, d.Get("branch").(string), d.Get("event").(string), d.Get("status").(string), d.Get("actor").(string), d.Get("head_sha").(string)))
	return nil
}

func dataSourceGiteaActionsJobs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaActionsJobsRead,
		Schema: mergeSchemaMaps(actionScopeSchema(), map[string]*schema.Schema{
			runIDField: {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "For `scope = \"repo\"`, list jobs for a specific run.",
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter jobs by status.",
			},
			"jobs": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The matching Actions jobs.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":           {Type: schema.TypeInt, Computed: true},
						"run_id":       {Type: schema.TypeInt, Computed: true},
						"run_url":      {Type: schema.TypeString, Computed: true},
						"run_attempt":  {Type: schema.TypeInt, Computed: true},
						"name":         {Type: schema.TypeString, Computed: true},
						"head_branch":  {Type: schema.TypeString, Computed: true},
						"head_sha":     {Type: schema.TypeString, Computed: true},
						"status":       {Type: schema.TypeString, Computed: true},
						"conclusion":   {Type: schema.TypeString, Computed: true},
						"url":          {Type: schema.TypeString, Computed: true},
						"html_url":     {Type: schema.TypeString, Computed: true},
						"created_at":   {Type: schema.TypeString, Computed: true},
						"started_at":   {Type: schema.TypeString, Computed: true},
						"completed_at": {Type: schema.TypeString, Computed: true},
						"runner_id":    {Type: schema.TypeInt, Computed: true},
						"runner_name":  {Type: schema.TypeString, Computed: true},
						"labels": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"steps": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name":         {Type: schema.TypeString, Computed: true},
									"number":       {Type: schema.TypeInt, Computed: true},
									"status":       {Type: schema.TypeString, Computed: true},
									"conclusion":   {Type: schema.TypeString, Computed: true},
									"started_at":   {Type: schema.TypeString, Computed: true},
									"completed_at": {Type: schema.TypeString, Computed: true},
								},
							},
						},
					},
				},
			},
			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		}),
		Description: "Lists Actions jobs for the selected scope.",
	}
}

func dataSourceGiteaActionsJobsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	cfg, err := resolveActionScope(d)
	if err != nil {
		return err
	}
	if err := requireVersion(client, ">= 1.26.0", "actions jobs"); err != nil {
		return err
	}

	runID := int64(d.Get(runIDField).(int))
	if cfg.Scope != actionScopeRepo && runID != 0 {
		return fmt.Errorf("%s is only supported for scope %q", runIDField, actionScopeRepo)
	}

	opt := gitea.ListRepoActionJobsOptions{
		Status: d.Get("status").(string),
	}
	jobs, err := collectPaginated(func(page int) ([]*gitea.ActionWorkflowJob, error) {
		opt.ListOptions = gitea.ListOptions{Page: page, PageSize: 100}
		response, callErr := listActionJobsByScope(client, cfg, runID, opt)
		if callErr != nil {
			return nil, callErr
		}
		return response.Jobs, nil
	})
	if err != nil {
		return err
	}

	if err := d.Set("jobs", flattenActionJobs(jobs)); err != nil {
		return fmt.Errorf("error setting jobs: %w", err)
	}
	d.Set("total_count", len(jobs))
	d.SetId(buildResourceID(cfg.Scope, cfg.Org, cfg.Owner, cfg.Repo, fmt.Sprintf("%d", runID), d.Get("status").(string)))
	return nil
}

func dataSourceGiteaRepositoryActionsWorkflows() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaRepositoryActionsWorkflowsRead,
		Schema: mergeSchemaMaps(repositoryIdentitySchema(), map[string]*schema.Schema{
			"workflows": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Repository workflow definitions.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":         {Type: schema.TypeString, Computed: true},
						"name":       {Type: schema.TypeString, Computed: true},
						"path":       {Type: schema.TypeString, Computed: true},
						"state":      {Type: schema.TypeString, Computed: true},
						"created_at": {Type: schema.TypeString, Computed: true},
						"updated_at": {Type: schema.TypeString, Computed: true},
						"url":        {Type: schema.TypeString, Computed: true},
						"html_url":   {Type: schema.TypeString, Computed: true},
						"badge_url":  {Type: schema.TypeString, Computed: true},
						"deleted_at": {Type: schema.TypeString, Computed: true},
					},
				},
			},
			"total_count": {Type: schema.TypeInt, Computed: true},
		}),
		Description: "Lists repository Actions workflows.",
	}
}

func dataSourceGiteaRepositoryActionsWorkflowsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	if err := requireVersion(client, ">= 1.25.0", "repository actions workflows"); err != nil {
		return err
	}
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repo := strings.ToLower(d.Get(repositoryNameField).(string))
	workflows, _, err := client.ListRepoActionWorkflows(owner, repo)
	if err != nil {
		return err
	}
	if err := d.Set("workflows", flattenActionWorkflows(workflows.Workflows)); err != nil {
		return fmt.Errorf("error setting workflows: %w", err)
	}
	d.Set("total_count", len(workflows.Workflows))
	d.SetId(buildResourceID(owner, repo))
	return nil
}

func dataSourceGiteaRepositoryActionsArtifacts() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaRepositoryActionsArtifactsRead,
		Schema: mergeSchemaMaps(repositoryIdentitySchema(), map[string]*schema.Schema{
			runIDField: {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "List artifacts for a specific run when set.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter artifacts by name.",
			},
			"artifacts": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The matching repository artifacts.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                   {Type: schema.TypeInt, Computed: true},
						"name":                 {Type: schema.TypeString, Computed: true},
						"size_in_bytes":        {Type: schema.TypeInt, Computed: true},
						"url":                  {Type: schema.TypeString, Computed: true},
						"archive_download_url": {Type: schema.TypeString, Computed: true},
						"expired":              {Type: schema.TypeBool, Computed: true},
						"workflow_run_id":      {Type: schema.TypeInt, Computed: true},
						"created_at":           {Type: schema.TypeString, Computed: true},
						"updated_at":           {Type: schema.TypeString, Computed: true},
						"expires_at":           {Type: schema.TypeString, Computed: true},
					},
				},
			},
			"total_count": {Type: schema.TypeInt, Computed: true},
		}),
		Description: "Lists repository Actions artifacts.",
	}
}

func dataSourceGiteaRepositoryActionsArtifactsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	if err := requireVersion(client, ">= 1.25.0", "repository actions artifacts"); err != nil {
		return err
	}
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repo := strings.ToLower(d.Get(repositoryNameField).(string))
	runID := int64(d.Get(runIDField).(int))
	name := d.Get("name").(string)

	artifacts, err := collectPaginated(func(page int) ([]*gitea.ActionArtifact, error) {
		opt := gitea.ListActionArtifactsOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 100},
			Name:        name,
		}
		var response *gitea.ActionArtifactsResponse
		var callErr error
		if runID != 0 {
			response, _, callErr = client.ListRepoActionRunArtifacts(owner, repo, runID, opt)
		} else {
			response, _, callErr = client.ListRepoActionArtifacts(owner, repo, opt)
		}
		if callErr != nil {
			return nil, callErr
		}
		return response.Artifacts, nil
	})
	if err != nil {
		return err
	}

	if err := d.Set("artifacts", flattenActionArtifacts(artifacts)); err != nil {
		return fmt.Errorf("error setting artifacts: %w", err)
	}
	d.Set("total_count", len(artifacts))
	d.SetId(buildResourceID(owner, repo, fmt.Sprintf("%d", runID), name))
	return nil
}

func resourceGiteaRepositoryActionsWorkflowState() *schema.Resource {
	return &schema.Resource{
		Create: resourceGiteaRepositoryActionsWorkflowStateCreate,
		Read:   resourceGiteaRepositoryActionsWorkflowStateRead,
		Update: resourceGiteaRepositoryActionsWorkflowStateUpdate,
		Delete: resourceGiteaRepositoryActionsWorkflowStateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				owner, repo, workflowID, err := parseThreePartID(d.Id(), repositoryOwnerField, repositoryNameField, workflowIDField)
				if err != nil {
					return nil, err
				}
				d.Set(repositoryOwnerField, owner)
				d.Set(repositoryNameField, repo)
				d.Set(workflowIDField, workflowID)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: mergeSchemaMaps(repositoryIdentitySchema(), map[string]*schema.Schema{
			workflowIDField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The workflow ID or file name.",
			},
			enabledField: {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether the workflow is enabled.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The workflow display name.",
			},
			"path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The workflow file path.",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The workflow state returned by Gitea.",
			},
			createdAtField: {Type: schema.TypeString, Computed: true},
			updatedAtField: {Type: schema.TypeString, Computed: true},
		}),
		Description: "`gitea_repository_actions_workflow_state` manages the enabled state of an existing repository workflow.\n\n" +
			"Deleting the resource only removes it from Terraform state and leaves the remote workflow unchanged.\n\n" +
			"Import expects the resource ID in the form `owner:repo:workflow_id`.",
	}
}

func resourceGiteaRepositoryActionsWorkflowStateCreate(d *schema.ResourceData, meta interface{}) error {
	if err := setRepositoryWorkflowState(d, meta); err != nil {
		return err
	}
	d.SetId(buildThreePartID(strings.ToLower(d.Get(repositoryOwnerField).(string)), strings.ToLower(d.Get(repositoryNameField).(string)), d.Get(workflowIDField).(string)))
	return resourceGiteaRepositoryActionsWorkflowStateRead(d, meta)
}

func resourceGiteaRepositoryActionsWorkflowStateRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	if err := requireVersion(client, ">= 1.25.0", "repository actions workflow state"); err != nil {
		return err
	}
	owner, repo, workflowID, err := parseThreePartID(d.Id(), repositoryOwnerField, repositoryNameField, workflowIDField)
	if err != nil {
		return err
	}
	workflow, resp, err := client.GetRepoActionWorkflow(owner, repo, workflowID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return err
	}
	d.Set(repositoryOwnerField, owner)
	d.Set(repositoryNameField, repo)
	d.Set(workflowIDField, workflowID)
	d.Set(enabledField, workflow.State != "disabled")
	d.Set("name", workflow.Name)
	d.Set("path", workflow.Path)
	d.Set("state", workflow.State)
	d.Set(createdAtField, timeToString(workflow.CreatedAt))
	d.Set(updatedAtField, timeToString(workflow.UpdatedAt))
	return nil
}

func resourceGiteaRepositoryActionsWorkflowStateUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := setRepositoryWorkflowState(d, meta); err != nil {
		return err
	}
	return resourceGiteaRepositoryActionsWorkflowStateRead(d, meta)
}

func resourceGiteaRepositoryActionsWorkflowStateDelete(d *schema.ResourceData, meta interface{}) error {
	_, _, _, err := parseThreePartID(d.Id(), repositoryOwnerField, repositoryNameField, workflowIDField)
	return err
}

func setRepositoryWorkflowState(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	if err := requireVersion(client, ">= 1.25.0", "repository actions workflow state"); err != nil {
		return err
	}
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repo := strings.ToLower(d.Get(repositoryNameField).(string))
	workflowID := d.Get(workflowIDField).(string)
	var err error
	if d.Get(enabledField).(bool) {
		_, err = client.EnableRepoActionWorkflow(owner, repo, workflowID)
	} else {
		_, err = client.DisableRepoActionWorkflow(owner, repo, workflowID)
	}
	return err
}

func listActionRunnersByScope(client *gitea.Client, cfg actionScopeConfig, opt gitea.ListActionRunnersOptions) (*gitea.ActionRunnersResponse, error) {
	switch cfg.Scope {
	case actionScopeAdmin:
		response, _, err := client.ListAdminActionRunners(opt)
		return response, err
	case actionScopeUser:
		response, _, err := client.ListUserActionRunners(opt)
		return response, err
	case actionScopeOrg:
		response, _, err := client.ListOrgActionRunners(cfg.Org, opt)
		return response, err
	case actionScopeRepo:
		response, _, err := client.ListRepoActionRunners(cfg.Owner, cfg.Repo, opt)
		return response, err
	default:
		return nil, fmt.Errorf("unsupported scope %q", cfg.Scope)
	}
}

func listActionRunsByScope(client *gitea.Client, cfg actionScopeConfig, opt gitea.ListRepoActionRunsOptions) (*gitea.ActionWorkflowRunsResponse, error) {
	switch cfg.Scope {
	case actionScopeAdmin:
		response, _, err := client.ListAdminActionRuns(opt)
		return response, err
	case actionScopeUser:
		response, _, err := client.ListUserActionRuns(opt)
		return response, err
	case actionScopeOrg:
		response, _, err := client.ListOrgActionRuns(cfg.Org, opt)
		return response, err
	case actionScopeRepo:
		response, _, err := client.ListRepoActionRuns(cfg.Owner, cfg.Repo, opt)
		return response, err
	default:
		return nil, fmt.Errorf("unsupported scope %q", cfg.Scope)
	}
}

func listActionJobsByScope(client *gitea.Client, cfg actionScopeConfig, runID int64, opt gitea.ListRepoActionJobsOptions) (*gitea.ActionWorkflowJobsResponse, error) {
	switch cfg.Scope {
	case actionScopeAdmin:
		response, _, err := client.ListAdminActionJobs(opt)
		return response, err
	case actionScopeUser:
		response, _, err := client.ListUserActionJobs(opt)
		return response, err
	case actionScopeOrg:
		response, _, err := client.ListOrgActionJobs(cfg.Org, opt)
		return response, err
	case actionScopeRepo:
		if runID != 0 {
			response, _, err := client.ListRepoActionRunJobs(cfg.Owner, cfg.Repo, runID, opt)
			return response, err
		}
		response, _, err := client.ListRepoActionJobs(cfg.Owner, cfg.Repo, opt)
		return response, err
	default:
		return nil, fmt.Errorf("unsupported scope %q", cfg.Scope)
	}
}

func dataSourceGiteaActionsRunnerRegistrationToken() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaActionsRunnerRegistrationTokenRead,
		Schema: mergeSchemaMaps(actionScopeSchema(), map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The runner registration token.",
			},
		}),
		Description: "Creates and returns an Actions runner registration token for the selected scope.",
	}
}

func dataSourceGiteaActionsRunnerRegistrationTokenRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	cfg, err := resolveActionScope(d)
	if err != nil {
		return err
	}
	if err := requireVersion(client, ">= 1.22.0", "actions runner registration token"); err != nil {
		return err
	}
	var token *gitea.RegistrationToken
	switch cfg.Scope {
	case actionScopeAdmin:
		token, _, err = client.CreateAdminActionRunnerRegistrationToken()
	case actionScopeUser:
		token, _, err = client.CreateUserActionRunnerRegistrationToken()
	case actionScopeOrg:
		token, _, err = client.CreateOrgActionRunnerRegistrationToken(cfg.Org)
	case actionScopeRepo:
		token, _, err = client.CreateRepoActionRunnerRegistrationToken(cfg.Owner, cfg.Repo)
	default:
		err = fmt.Errorf("unsupported scope %q", cfg.Scope)
	}
	if err != nil {
		return err
	}
	d.Set("token", token.Token)
	d.SetId(buildResourceID("runner_registration_token", cfg.Scope, cfg.Org, cfg.Owner, cfg.Repo))
	return nil
}

func dataSourceGiteaRepositoryActionsArtifact() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaRepositoryActionsArtifactRead,
		Schema: mergeSchemaMaps(repositoryIdentitySchema(), map[string]*schema.Schema{
			"artifact_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The artifact ID.",
			},
			"name":                 {Type: schema.TypeString, Computed: true},
			"size_in_bytes":        {Type: schema.TypeInt, Computed: true},
			"url":                  {Type: schema.TypeString, Computed: true},
			"archive_download_url": {Type: schema.TypeString, Computed: true},
			"expired":              {Type: schema.TypeBool, Computed: true},
			"workflow_run_id":      {Type: schema.TypeInt, Computed: true},
			createdAtField:         {Type: schema.TypeString, Computed: true},
			updatedAtField:         {Type: schema.TypeString, Computed: true},
			"expires_at":           {Type: schema.TypeString, Computed: true},
		}),
		Description: "Fetches metadata for a single repository Actions artifact.",
	}
}

func dataSourceGiteaRepositoryActionsArtifactRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	if err := requireVersion(client, ">= 1.25.0", "repository actions artifact"); err != nil {
		return err
	}
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repo := strings.ToLower(d.Get(repositoryNameField).(string))
	artifactID := int64(d.Get("artifact_id").(int))

	artifact, _, err := client.GetRepoActionArtifact(owner, repo, artifactID)
	if err != nil {
		return err
	}
	runID := int64(0)
	if artifact.WorkflowRun != nil {
		runID = artifact.WorkflowRun.ID
	}
	d.Set("name", artifact.Name)
	d.Set("size_in_bytes", int(artifact.SizeInBytes))
	d.Set("url", artifact.URL)
	d.Set("archive_download_url", artifact.ArchiveDownloadURL)
	d.Set("expired", artifact.Expired)
	d.Set("workflow_run_id", int(runID))
	d.Set(createdAtField, timeToString(artifact.CreatedAt))
	d.Set(updatedAtField, timeToString(artifact.UpdatedAt))
	d.Set("expires_at", timeToString(artifact.ExpiresAt))
	d.SetId(buildResourceID(owner, repo, fmt.Sprintf("%d", artifactID)))
	return nil
}
