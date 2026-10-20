# Contributing to Shrike

Thank you for your interest in contributing to Shrike! This document provides guidelines for contributing to the project.

## Ways to Contribute

- **Report bugs**: Open an issue describing the bug and how to reproduce it
- **Suggest features**: Open an issue describing the feature and its use case
- **Add detection rules**: Implement new patterns for agent or workflow vulnerabilities
- **Improve documentation**: Fix typos, clarify explanations, add examples
- **Submit fixes**: Fix bugs or implement features via pull requests

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git

### Clone and Build

```bash
git clone https://github.com/chelate/shrike.git
cd shrike
go mod download
go build -o shrike .
```

### Project Structure

```
shrike/
├── cmd/              # CLI commands (scan, etc.)
├── internal/
│   ├── collector/    # Workflow collection (local, remote, org)
│   ├── parser/       # YAML parsing and workflow modeling
│   ├── detector/     # Detection rules and evaluation engine
│   ├── reporter/     # Output formatting (terminal, JSON, SARIF, markdown)
│   └── models/       # Data models and constants
├── main.go           # Entry point
└── action.yml        # GitHub Action definition
```

## Adding New Detection Rules

Detection rules live in `internal/detector/rules.go`. Each rule implements the `Rule` interface:

```go
type Rule interface {
    Evaluate(ctx *EvalContext) *Finding
    ID() string
}
```

### Example: Adding a New Rule

1. Create a new rule struct in `internal/detector/rules.go`:

```go
type MyNewRule struct{}

func (r *MyNewRule) ID() string {
    return "my-new-pattern"
}

func (r *MyNewRule) Evaluate(ctx *EvalContext) *Finding {
    // Check if the pattern applies
    if !shouldDetect(ctx) {
        return nil
    }

    // Determine severity based on legs present
    severity := "medium"
    if ctx.IsOutsiderTriggerable() {
        severity = "high"
    }

    return &Finding{
        ID:             r.ID(),
        Severity:       severity,
        Title:          "Brief description",
        Description:    "Detailed explanation",
        Triggerability: ctx.GetTriggerDescription(),
        Impact:         "What can be compromised",
        Location:       ctx.Location(),
        Prerequisites:  []string{"List", "of", "conditions"},
        RuleIDs:        []string{r.ID()},
    }
}
```

2. Add the rule to `allRules()` in `rules.go`:

```go
func allRules() []Rule {
    return []Rule{
        // ... existing rules
        &MyNewRule{},
    }
}
```

3. Test your rule with example workflows in `test-repo/.github/workflows/`

## Adding New Agent Actions

To add support for a new AI agent action:

1. Add the action to `KnownAgentActions` in `internal/models/workflow.go`:

```go
var KnownAgentActions = map[string]AgentActionInfo{
    // ... existing actions
    "vendor/new-agent-action": {
        Name:        "New Agent Name",
        Vendor:      "Vendor Name",
        PromptField: "prompt",      // Field containing the prompt
        ArgsField:   "args",         // Field containing tool/args config
        GateField:   "allowed_users", // Field for authorization gate
    },
}
```

2. If the agent can be invoked via CLI (not just as an action), add patterns to `AgentCLIPatterns`:

```go
var AgentCLIPatterns = []string{
    // ... existing patterns
    "npx @vendor/new-agent",
}
```

## Testing

### Manual Testing

```bash
# Build
go build -o shrike .

# Test on local workflows
./shrike scan test-repo

# Test specific output formats
./shrike scan test-repo --json
./shrike scan test-repo --sarif
./shrike scan test-repo --verbose
```

### Test Cases

Add test workflows to `test-repo/.github/workflows/` that demonstrate:
- The vulnerability your rule detects
- Edge cases
- False positive scenarios (should NOT be detected)

## Pull Request Guidelines

1. **Create an issue first** for major changes to discuss the approach
2. **Keep PRs focused**: One feature or fix per PR
3. **Write clear commit messages**: Explain what and why, not just what
4. **Test your changes**: Ensure existing tests pass and add new tests
5. **Update documentation**: Update README.md if adding user-facing features
6. **Follow Go conventions**: Run `go fmt` and `go vet`

### PR Template

```markdown
## Description
Brief description of what this PR does

## Motivation
Why is this change needed? What problem does it solve?

## Changes
- List of changes made
- Including new files or modified logic

## Testing
How was this tested? Include example commands and expected output

## Related Issues
Fixes #issue_number
```

## Code Style

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable names
- Add comments for non-obvious logic
- Keep functions focused and reasonably sized
- Prefer clarity over cleverness

## Documentation

- Update README.md for user-facing changes
- Add inline comments for complex detection logic
- Include examples in documentation

## Questions?

Open an issue or discussion on GitHub if you have questions about:
- How to implement a feature
- Whether a contribution would be accepted
- Technical details about the codebase

## License

By contributing to Shrike, you agree that your contributions will be licensed under the MIT License.
