package models

// Workflow represents a raw workflow file
type Workflow struct {
	Path       string // Relative path from repo root
	RepoOwner  string // For remote scans
	RepoName   string // For remote scans
	Content    []byte // Raw YAML content
	Name       string // Workflow name from the file
}

// WorkflowYAML represents the parsed YAML structure
type WorkflowYAML struct {
	Name string                 `yaml:"name"`
	On   interface{}            `yaml:"on"` // Can be string, []string, or map
	Jobs map[string]JobYAML     `yaml:"jobs"`
	Permissions interface{}     `yaml:"permissions"` // Can be string or map
	Env map[string]string       `yaml:"env"`
}

// JobYAML represents a job in the workflow
type JobYAML struct {
	Name        string              `yaml:"name"`
	RunsOn      interface{}         `yaml:"runs-on"` // string or []string
	Steps       []StepYAML          `yaml:"steps"`
	Permissions interface{}         `yaml:"permissions"`
	Env         map[string]string   `yaml:"env"`
	If          string              `yaml:"if"`
	Needs       interface{}         `yaml:"needs"` // string or []string
}

// StepYAML represents a step in a job
type StepYAML struct {
	Name string                 `yaml:"name"`
	Uses string                 `yaml:"uses"`
	Run  string                 `yaml:"run"`
	With map[string]interface{} `yaml:"with"`
	Env  map[string]string      `yaml:"env"`
	If   string                 `yaml:"if"`
}

// Known agent actions
var KnownAgentActions = map[string]AgentActionInfo{
	"anthropics/claude-code-action": {
		Name:        "Claude Code Action",
		Vendor:      "Anthropic",
		PromptField: "prompt",
		ArgsField:   "claude_args",
		GateField:   "allowed_non_write_users",
	},
	"google/gemini-cli-action": {
		Name:        "Gemini CLI Action",
		Vendor:      "Google",
		PromptField: "direct_prompt",
		ArgsField:   "gemini_args",
	},
	"openai/codex-action": {
		Name:        "OpenAI Codex Action",
		Vendor:      "OpenAI",
		PromptField: "prompt",
		ArgsField:   "tools",
	},
}

// Agent CLI patterns that might run in a step
var AgentCLIPatterns = []string{
	"npx @google/gemini-cli",
	"claude -p",
	"claude --prompt",
	"aider",
	"cursor",
}

type AgentActionInfo struct {
	Name        string
	Vendor      string
	PromptField string // Field name that contains the prompt
	ArgsField   string // Field name that contains tool/args config
	GateField   string // Field name for authorization gate (e.g., allowed_non_write_users)
}

// Trigger types
type TriggerType string

const (
	TriggerIssues              TriggerType = "issues"
	TriggerIssueComment        TriggerType = "issue_comment"
	TriggerPullRequest         TriggerType = "pull_request"
	TriggerPullRequestTarget   TriggerType = "pull_request_target"
	TriggerWorkflowRun         TriggerType = "workflow_run"
	TriggerPush                TriggerType = "push"
	TriggerWorkflowDispatch    TriggerType = "workflow_dispatch"
	TriggerSchedule            TriggerType = "schedule"
	TriggerRelease             TriggerType = "release"
	TriggerOther               TriggerType = "other"
)

// Outsider-triggerable events
var OutsiderTriggers = map[TriggerType]bool{
	TriggerIssues:            true,
	TriggerIssueComment:      true,
	TriggerPullRequestTarget: true,
	// pull_request is triggerable if from forks (needs deeper check)
}

// Untrusted GitHub context expressions
var UntrustedExpressions = []string{
	"github.event.issue.title",
	"github.event.issue.body",
	"github.event.comment.body",
	"github.event.pull_request.title",
	"github.event.pull_request.body",
	"github.head_ref",
	"github.event.head_commit.message",
}

// Dangerous permissions
type PermissionLevel string

const (
	PermissionWrite PermissionLevel = "write"
	PermissionRead  PermissionLevel = "read"
	PermissionNone  PermissionLevel = "none"
)

type Permissions struct {
	Contents      PermissionLevel
	PullRequests  PermissionLevel
	Issues        PermissionLevel
	Packages      PermissionLevel
	Deployments   PermissionLevel
	DefaultWrite  bool // true if no permissions block (inherits write)
}
