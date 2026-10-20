package detector

import (
	"fmt"
	"strings"

	"github.com/chelate/shrike/internal/models"
	"github.com/chelate/shrike/internal/parser"
)

type Detector struct {
	rules []Rule
}

func New() *Detector {
	return &Detector{
		rules: allRules(),
	}
}

// Finding represents a detected security risk
type Finding struct {
	ID              string
	Severity        string // critical, high, medium, low
	Title           string
	Description     string
	Triggerability  string // Who can trigger and how
	Impact          string // What it can reach
	Location        Location
	VulnerableLines []string
	Prerequisites   []string
	Repro           string
	RuleIDs         []string // Which rules matched
}

type Location struct {
	RepoOwner    string
	RepoName     string
	WorkflowPath string
	WorkflowName string
	JobID        string
	JobName      string
	StepIndex    int
	StepName     string
}

func (l Location) String() string {
	parts := []string{l.WorkflowPath, l.JobID}
	if l.StepName != "" {
		parts = append(parts, fmt.Sprintf("step %d (%s)", l.StepIndex, l.StepName))
	} else {
		parts = append(parts, fmt.Sprintf("step %d", l.StepIndex))
	}
	return strings.Join(parts, " → ")
}

// Evaluate runs all detection rules against parsed workflows
func (d *Detector) Evaluate(workflows []*parser.ParsedWorkflow) []*Finding {
	var findings []*Finding

	for _, wf := range workflows {
		for _, job := range wf.Jobs {
			for stepIdx, step := range job.Steps {
				// Build context for this step
				ctx := &EvalContext{
					Workflow:  wf,
					Job:       job,
					Step:      step,
					StepIndex: stepIdx,
				}

				// Run rules
				for _, rule := range d.rules {
					if finding := rule.Evaluate(ctx); finding != nil {
						findings = append(findings, finding)
					}
				}
			}
		}
	}

	// Deduplicate and rank findings
	findings = deduplicateFindings(findings)
	rankFindings(findings)

	return findings
}

// EvalContext provides context for rule evaluation
type EvalContext struct {
	Workflow  *parser.ParsedWorkflow
	Job       *parser.ParsedJob
	Step      *parser.ParsedStep
	StepIndex int
}

// Helper methods for context

func (ctx *EvalContext) IsOutsiderTriggerable() bool {
	for _, trigger := range ctx.Workflow.Triggers {
		if models.OutsiderTriggers[trigger.Type] {
			return true
		}
		// Check pull_request with fork support
		if trigger.Type == models.TriggerPullRequest {
			// pull_request from forks is triggerable unless explicitly limited
			if filter := trigger.Filter; filter != nil {
				if branches, ok := filter["branches"]; ok {
					// If branches are specified but doesn't exclude forks, it's triggerable
					_ = branches
					return true
				}
			}
			return true // Default: pull_request accepts forks
		}
	}
	return false
}

func (ctx *EvalContext) GetTriggerDescription() string {
	var triggers []string
	for _, t := range ctx.Workflow.Triggers {
		triggers = append(triggers, string(t.Type))
	}
	if len(triggers) == 0 {
		return "unknown trigger"
	}
	return strings.Join(triggers, ", ")
}

func (ctx *EvalContext) HasAuthGate() bool {
	if !ctx.Step.IsAgentStep {
		return false
	}

	// Check for allowed_non_write_users or similar gates
	if ctx.Step.AgentInfo != nil && ctx.Step.AgentInfo.GateField != "" {
		if val, ok := ctx.Step.With[ctx.Step.AgentInfo.GateField]; ok {
			// Check if it's a wildcard (anyone can trigger)
			if strVal, ok := val.(string); ok {
				if strVal == "*" {
					return false // Wildcard = no gate
				}
				return true // Some restriction exists
			}
		}
	}

	// Check for author_association in step condition
	if strings.Contains(ctx.Step.Condition, "author_association") {
		if strings.Contains(ctx.Step.Condition, "OWNER") ||
			strings.Contains(ctx.Step.Condition, "MEMBER") ||
			strings.Contains(ctx.Step.Condition, "COLLABORATOR") {
			return true
		}
	}

	// Check job-level condition
	if strings.Contains(ctx.Job.Condition, "author_association") {
		return true
	}

	return false
}

func (ctx *EvalContext) GetUntrustedInputs() []string {
	var inputs []string

	// Get the prompt field for agent steps
	var promptValue string
	if ctx.Step.IsAgentStep && ctx.Step.AgentInfo != nil {
		if val, ok := ctx.Step.With[ctx.Step.AgentInfo.PromptField]; ok {
			if strVal, ok := val.(string); ok {
				promptValue = strVal
			}
		}
	}

	// Also check run commands for non-agent steps
	searchText := promptValue + " " + ctx.Step.Run

	// Check for untrusted expressions
	for _, expr := range models.UntrustedExpressions {
		if strings.Contains(searchText, expr) {
			inputs = append(inputs, expr)
		}
	}

	return inputs
}

func (ctx *EvalContext) HasDangerousPermissions() bool {
	perms := ctx.Job.Permissions
	if perms == nil {
		return false
	}

	return perms.DefaultWrite ||
		perms.Contents == models.PermissionWrite ||
		perms.PullRequests == models.PermissionWrite ||
		perms.Packages == models.PermissionWrite ||
		perms.Deployments == models.PermissionWrite
}

func (ctx *EvalContext) GetPermissionsDescription() string {
	perms := ctx.Job.Permissions
	if perms == nil {
		return "default permissions"
	}

	var parts []string
	if perms.DefaultWrite {
		parts = append(parts, "inherited write")
	}
	if perms.Contents == models.PermissionWrite {
		parts = append(parts, "contents:write")
	}
	if perms.PullRequests == models.PermissionWrite {
		parts = append(parts, "pull-requests:write")
	}
	if perms.Packages == models.PermissionWrite {
		parts = append(parts, "packages:write")
	}

	if len(parts) == 0 {
		return "read-only permissions"
	}
	return strings.Join(parts, ", ")
}

func (ctx *EvalContext) HasDangerousSecrets() bool {
	// Check job-level env for secrets
	for key := range ctx.Job.Env {
		if isDangerousSecret(key) {
			return true
		}
	}
	// Check step-level env
	for key := range ctx.Step.Env {
		if isDangerousSecret(key) {
			return true
		}
	}
	return false
}

func isDangerousSecret(key string) bool {
	keyLower := strings.ToLower(key)
	dangerous := []string{
		"npm_token", "pypi_token", "aws_access_key", "aws_secret",
		"gcp_credentials", "azure_credentials", "docker_password",
		"github_token", "deploy_key", "api_key", "secret_key",
	}
	for _, d := range dangerous {
		if strings.Contains(keyLower, d) {
			return true
		}
	}
	return false
}

func (ctx *EvalContext) GetAgentTools() []string {
	if !ctx.Step.IsAgentStep {
		return nil
	}

	var tools []string

	// Check claude_args for tools
	if argsVal, ok := ctx.Step.With["claude_args"]; ok {
		if argsStr, ok := argsVal.(string); ok {
			if strings.Contains(argsStr, "Bash") {
				tools = append(tools, "Bash")
			}
			if strings.Contains(argsStr, "Write") {
				tools = append(tools, "Write")
			}
			if strings.Contains(argsStr, "Edit") {
				tools = append(tools, "Edit")
			}
		}
	}

	// Check for unscoped tool access (Gemini, etc.)
	if ctx.Step.AgentInfo != nil && ctx.Step.AgentInfo.Vendor == "Google" {
		// Gemini without explicit tool restrictions
		if _, hasToolsCore := ctx.Step.With["tools.core"]; !hasToolsCore {
			tools = append(tools, "unrestricted shell")
		}
	}

	return tools
}

func (ctx *EvalContext) Location() Location {
	return Location{
		RepoOwner:    ctx.Workflow.Raw.RepoOwner,
		RepoName:     ctx.Workflow.Raw.RepoName,
		WorkflowPath: ctx.Workflow.Raw.Path,
		WorkflowName: ctx.Workflow.Name,
		JobID:        ctx.Job.ID,
		JobName:      ctx.Job.Name,
		StepIndex:    ctx.StepIndex,
		StepName:     ctx.Step.Name,
	}
}

func deduplicateFindings(findings []*Finding) []*Finding {
	// Simple dedup by location for now
	seen := make(map[string]bool)
	var unique []*Finding

	for _, f := range findings {
		key := f.Location.String() + f.ID
		if !seen[key] {
			seen[key] = true
			unique = append(unique, f)
		}
	}

	return unique
}

func rankFindings(findings []*Finding) {
	// Sort by severity: critical > high > medium > low
	// This is a simple bubble sort - good enough for typical finding counts
	severityRank := map[string]int{
		"critical": 4,
		"high":     3,
		"medium":   2,
		"low":      1,
	}

	for i := 0; i < len(findings); i++ {
		for j := i + 1; j < len(findings); j++ {
			if severityRank[findings[i].Severity] < severityRank[findings[j].Severity] {
				findings[i], findings[j] = findings[j], findings[i]
			}
		}
	}
}
