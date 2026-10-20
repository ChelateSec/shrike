package parser

import (
	"fmt"
	"strings"

	"github.com/chelate/shrike/internal/models"
	"gopkg.in/yaml.v3"
)

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

// ParsedWorkflow is the analyzed representation after parsing
type ParsedWorkflow struct {
	Raw       *models.Workflow
	Name      string
	Triggers  []Trigger
	Jobs      []*ParsedJob
	Permissions *models.Permissions
}

type Trigger struct {
	Type   models.TriggerType
	Events []string // Specific events for the trigger
	Filter map[string]interface{} // branches, types, etc.
}

type ParsedJob struct {
	ID          string
	Name        string
	Steps       []*ParsedStep
	Permissions *models.Permissions
	Env         map[string]string
	Condition   string // The 'if' field
	RunsOn      string
}

type ParsedStep struct {
	Name         string
	Uses         string // Action reference
	Run          string // Shell command
	With         map[string]interface{}
	Env          map[string]string
	Condition    string
	IsAgentStep  bool
	AgentInfo    *models.AgentActionInfo
	LineNumber   int
}

func (p *Parser) Parse(wf *models.Workflow) (*ParsedWorkflow, error) {
	var workflowYAML models.WorkflowYAML
	if err := yaml.Unmarshal(wf.Content, &workflowYAML); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	parsed := &ParsedWorkflow{
		Raw:  wf,
		Name: workflowYAML.Name,
	}

	// Parse triggers
	parsed.Triggers = p.parseTriggers(workflowYAML.On)

	// Parse workflow-level permissions
	parsed.Permissions = p.parsePermissions(workflowYAML.Permissions)

	// Parse jobs
	for jobID, jobYAML := range workflowYAML.Jobs {
		parsedJob := p.parseJob(jobID, jobYAML, parsed.Permissions)
		parsed.Jobs = append(parsed.Jobs, parsedJob)
	}

	return parsed, nil
}

func (p *Parser) parseTriggers(on interface{}) []Trigger {
	var triggers []Trigger

	switch v := on.(type) {
	case string:
		// Single trigger: on: push
		triggers = append(triggers, Trigger{
			Type: p.normalizeTriggerType(v),
		})

	case []interface{}:
		// List of triggers: on: [push, pull_request]
		for _, item := range v {
			if triggerStr, ok := item.(string); ok {
				triggers = append(triggers, Trigger{
					Type: p.normalizeTriggerType(triggerStr),
				})
			}
		}

	case map[string]interface{}:
		// Map of triggers with config
		for triggerName, config := range v {
			trigger := Trigger{
				Type: p.normalizeTriggerType(triggerName),
			}

			// Parse trigger configuration
			if configMap, ok := config.(map[string]interface{}); ok {
				trigger.Filter = configMap

				// Extract specific events for pull_request, issues, etc.
				if types, ok := configMap["types"].([]interface{}); ok {
					for _, t := range types {
						if typeStr, ok := t.(string); ok {
							trigger.Events = append(trigger.Events, typeStr)
						}
					}
				}
			}

			triggers = append(triggers, trigger)
		}
	}

	return triggers
}

func (p *Parser) normalizeTriggerType(s string) models.TriggerType {
	switch s {
	case "issues":
		return models.TriggerIssues
	case "issue_comment":
		return models.TriggerIssueComment
	case "pull_request":
		return models.TriggerPullRequest
	case "pull_request_target":
		return models.TriggerPullRequestTarget
	case "workflow_run":
		return models.TriggerWorkflowRun
	case "push":
		return models.TriggerPush
	case "workflow_dispatch":
		return models.TriggerWorkflowDispatch
	case "schedule":
		return models.TriggerSchedule
	case "release":
		return models.TriggerRelease
	default:
		return models.TriggerOther
	}
}

func (p *Parser) parsePermissions(perms interface{}) *models.Permissions {
	permissions := &models.Permissions{}

	switch v := perms.(type) {
	case nil:
		// No permissions block = inherits repo defaults (usually write for workflows)
		permissions.DefaultWrite = true

	case string:
		// Shorthand: permissions: read-all or write-all
		if v == "write-all" {
			permissions.Contents = models.PermissionWrite
			permissions.PullRequests = models.PermissionWrite
			permissions.Issues = models.PermissionWrite
			permissions.DefaultWrite = true
		}

	case map[string]interface{}:
		// Detailed permissions map
		if contents, ok := v["contents"].(string); ok {
			permissions.Contents = p.normalizePermissionLevel(contents)
		}
		if prs, ok := v["pull-requests"].(string); ok {
			permissions.PullRequests = p.normalizePermissionLevel(prs)
		}
		if issues, ok := v["issues"].(string); ok {
			permissions.Issues = p.normalizePermissionLevel(issues)
		}
		if packages, ok := v["packages"].(string); ok {
			permissions.Packages = p.normalizePermissionLevel(packages)
		}
		if deployments, ok := v["deployments"].(string); ok {
			permissions.Deployments = p.normalizePermissionLevel(deployments)
		}
	}

	return permissions
}

func (p *Parser) normalizePermissionLevel(s string) models.PermissionLevel {
	switch s {
	case "write":
		return models.PermissionWrite
	case "read":
		return models.PermissionRead
	case "none":
		return models.PermissionNone
	default:
		return models.PermissionNone
	}
}

func (p *Parser) parseJob(jobID string, job models.JobYAML, workflowPerms *models.Permissions) *ParsedJob {
	parsed := &ParsedJob{
		ID:        jobID,
		Name:      job.Name,
		Env:       job.Env,
		Condition: job.If,
	}

	// Parse runs-on
	switch v := job.RunsOn.(type) {
	case string:
		parsed.RunsOn = v
	case []interface{}:
		if len(v) > 0 {
			if s, ok := v[0].(string); ok {
				parsed.RunsOn = s
			}
		}
	}

	// Job-level permissions override workflow-level
	if job.Permissions != nil {
		parsed.Permissions = p.parsePermissions(job.Permissions)
	} else {
		parsed.Permissions = workflowPerms
	}

	// Parse steps
	for i, step := range job.Steps {
		parsedStep := p.parseStep(step, i)
		parsed.Steps = append(parsed.Steps, parsedStep)
	}

	return parsed
}

func (p *Parser) parseStep(step models.StepYAML, lineNum int) *ParsedStep {
	parsed := &ParsedStep{
		Name:       step.Name,
		Uses:       step.Uses,
		Run:        step.Run,
		With:       step.With,
		Env:        step.Env,
		Condition:  step.If,
		LineNumber: lineNum,
	}

	// Check if this is an agent step
	if step.Uses != "" {
		parsed.IsAgentStep, parsed.AgentInfo = p.isAgentAction(step.Uses)
	} else if step.Run != "" {
		// Check for agent CLI patterns in run commands
		parsed.IsAgentStep, parsed.AgentInfo = p.isAgentCLI(step.Run)
	}

	return parsed
}

func (p *Parser) isAgentAction(uses string) (bool, *models.AgentActionInfo) {
	// Extract action name from uses (format: owner/repo@version or owner/repo/path@version)
	parts := strings.Split(uses, "@")
	if len(parts) == 0 {
		return false, nil
	}

	actionPath := parts[0]

	// Check against known agent actions
	for actionKey, info := range models.KnownAgentActions {
		if strings.HasPrefix(actionPath, actionKey) {
			infoCopy := info
			return true, &infoCopy
		}
	}

	return false, nil
}

func (p *Parser) isAgentCLI(run string) (bool, *models.AgentActionInfo) {
	runLower := strings.ToLower(run)

	for _, pattern := range models.AgentCLIPatterns {
		if strings.Contains(runLower, pattern) {
			// Generic CLI agent info
			return true, &models.AgentActionInfo{
				Name:   "Agent CLI",
				Vendor: "Unknown",
			}
		}
	}

	return false, nil
}

// Helper: check if a string contains GitHub expression interpolation
func ContainsExpression(s string) bool {
	return strings.Contains(s, "${{") || strings.Contains(s, "${")
}

// Helper: extract expressions from a string
func ExtractExpressions(s string) []string {
	var expressions []string
	for {
		start := strings.Index(s, "${{")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "}}")
		if end == -1 {
			break
		}
		end += start
		expr := strings.TrimSpace(s[start+3 : end])
		expressions = append(expressions, expr)
		s = s[end+2:]
	}
	return expressions
}
