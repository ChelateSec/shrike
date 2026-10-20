package collector

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chelate/shrike/internal/models"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type TargetType int

const (
	TargetLocal TargetType = iota
	TargetRepo
	TargetOrg
)

type ScanTarget struct {
	Type            TargetType
	Path            string // For local
	Owner           string // For repo
	Repo            string // For repo
	Org             string // For org
	Concurrency     int    // For org
	IncludeArchived bool   // For org
}

func (st ScanTarget) RepoCount() int {
	if st.Type == TargetLocal || st.Type == TargetRepo {
		return 1
	}
	// For org, we'll update this after scanning
	return 0
}

type Collector struct {
	client *github.Client
}

func New(token string) *Collector {
	var client *github.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(context.Background(), ts)
		client = github.NewClient(tc)
	}
	return &Collector{client: client}
}

func (c *Collector) Collect(ctx context.Context, target ScanTarget) ([]*models.Workflow, error) {
	switch target.Type {
	case TargetLocal:
		return c.collectLocal(target.Path)
	case TargetRepo:
		return c.collectRepo(ctx, target.Owner, target.Repo)
	case TargetOrg:
		return c.collectOrg(ctx, target.Org, target.Concurrency, target.IncludeArchived)
	default:
		return nil, fmt.Errorf("unknown target type")
	}
}

func (c *Collector) collectLocal(path string) ([]*models.Workflow, error) {
	workflowDir := filepath.Join(path, ".github", "workflows")

	// Check if directory exists
	if _, err := os.Stat(workflowDir); os.IsNotExist(err) {
		return nil, nil // No workflows directory
	}

	var workflows []*models.Workflow

	err := filepath.WalkDir(workflowDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Only YAML files
		ext := filepath.Ext(path)
		if ext != ".yml" && ext != ".yaml" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Get relative path from repo root
		relPath, err := filepath.Rel(filepath.Dir(filepath.Dir(filepath.Dir(workflowDir))), path)
		if err != nil {
			relPath = path
		}

		workflows = append(workflows, &models.Workflow{
			Path:    relPath,
			Content: content,
			Name:    filepath.Base(path),
		})

		return nil
	})

	return workflows, err
}

func (c *Collector) collectRepo(ctx context.Context, owner, repo string) ([]*models.Workflow, error) {
	if c.client == nil {
		return nil, fmt.Errorf("GitHub client not initialized (GITHUB_TOKEN required)")
	}

	// List workflow files in .github/workflows
	_, dirContent, _, err := c.client.Repositories.GetContents(ctx, owner, repo, ".github/workflows", nil)
	if err != nil {
		// 404 is okay - just no workflows
		if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == 404 {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}

	var workflows []*models.Workflow

	for _, item := range dirContent {
		if item.Type == nil || *item.Type != "file" {
			continue
		}
		name := item.GetName()
		ext := filepath.Ext(name)
		if ext != ".yml" && ext != ".yaml" {
			continue
		}

		// Fetch file content
		fileContent, _, _, err := c.client.Repositories.GetContents(ctx, owner, repo, item.GetPath(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", name, err)
		}

		content, err := fileContent.GetContent()
		if err != nil {
			return nil, fmt.Errorf("failed to decode %s: %w", name, err)
		}

		workflows = append(workflows, &models.Workflow{
			Path:      item.GetPath(),
			RepoOwner: owner,
			RepoName:  repo,
			Content:   []byte(content),
			Name:      name,
		})
	}

	return workflows, nil
}

func (c *Collector) collectOrg(ctx context.Context, org string, concurrency int, includeArchived bool) ([]*models.Workflow, error) {
	if c.client == nil {
		return nil, fmt.Errorf("GitHub client not initialized (GITHUB_TOKEN required)")
	}

	// List all repos in org
	var allRepos []*github.Repository
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := c.client.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		for _, repo := range repos {
			// Skip archived repos unless requested
			if repo.GetArchived() && !includeArchived {
				continue
			}
			allRepos = append(allRepos, repo)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage

		// Rate limit check with backoff
		if err := c.checkRateLimit(ctx); err != nil {
			return nil, err
		}
	}

	if len(allRepos) == 0 {
		return nil, nil
	}

	// Scan repos concurrently
	type result struct {
		workflows []*models.Workflow
		err       error
	}

	sem := make(chan struct{}, concurrency)
	results := make(chan result, len(allRepos))
	var wg sync.WaitGroup

	for _, repo := range allRepos {
		wg.Add(1)
		go func(r *github.Repository) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			workflows, err := c.collectRepo(ctx, r.GetOwner().GetLogin(), r.GetName())
			results <- result{workflows: workflows, err: err}

			// Small delay to respect rate limits
			time.Sleep(100 * time.Millisecond)
		}(repo)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var allWorkflows []*models.Workflow
	for res := range results {
		if res.err != nil {
			// Log but don't fail entire scan
			fmt.Fprintf(os.Stderr, "Warning: %v\n", res.err)
			continue
		}
		allWorkflows = append(allWorkflows, res.workflows...)
	}

	return allWorkflows, nil
}

func (c *Collector) checkRateLimit(ctx context.Context) error {
	limits, _, err := c.client.RateLimits(ctx)
	if err != nil {
		return err
	}

	core := limits.GetCore()
	if core.Remaining < 100 {
		resetTime := core.Reset.Time
		waitDuration := time.Until(resetTime)
		if waitDuration > 0 {
			fmt.Fprintf(os.Stderr, "Rate limit low (%d remaining), waiting until %s...\n",
				core.Remaining, resetTime.Format(time.RFC3339))
			time.Sleep(waitDuration)
		}
	}

	return nil
}

// Helper to detect if string contains workflow expression interpolation
func ContainsExpression(s string) bool {
	return strings.Contains(s, "${{") || strings.Contains(s, "${")
}
