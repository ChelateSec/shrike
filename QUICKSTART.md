# Shrike Quick Start Guide

Get started with Shrike in 2 minutes.

## Installation

### Option 1: Build from Source (Recommended for Development)

```bash
git clone https://github.com/chelate/shrike.git
cd shrike
go build -o shrike .
```

### Option 2: Download Pre-built Binary

```bash
# Linux
curl -sL https://github.com/chelate/shrike/releases/latest/download/shrike-linux-amd64 -o shrike
chmod +x shrike

# macOS
curl -sL https://github.com/chelate/shrike/releases/latest/download/shrike-darwin-amd64 -o shrike
chmod +x shrike
```

### Option 3: Go Install

```bash
go install github.com/chelate/shrike@latest
```

## First Scan

### Scan Your Current Repository

```bash
# From the repository root
shrike scan .
```

**Example output:**
```
Scanned 5 workflow(s) across 1 repository

[1] CRITICAL - AI agent triggerable by any GitHub user without authorization
    Triggerable: Any GitHub user via issues, no authorization gate
    Impact: Agent execution with contents:write, tools: Bash
    Location: .github/workflows/ai-assist.yml → respond → step 2

Found 1 finding(s)
```

### Understanding the Output

Each finding shows:
- **Severity**: CRITICAL, HIGH, MEDIUM, or LOW
- **Triggerable by**: Who can exploit this and how
- **Impact**: What they can reach (permissions, tools, secrets)
- **Location**: Exact file → job → step

### No Findings? You're Good!

```
No triggerable agent or workflow risks found in 5 workflows
```

This means Shrike found no complete attack paths in your workflows.

## Common Use Cases

### 1. Security Audit

```bash
# Scan and generate detailed report
shrike scan . --verbose --report security-audit.md
```

### 2. CI/CD Gate

```bash
# Fail build on critical/high severity
shrike scan . --fail-on high
```

### 3. Organization-wide Scan

```bash
# Scan all repos in your org
export GITHUB_TOKEN=ghp_your_token_here
shrike scan --github-org your-org-name
```

### 4. Export for Further Analysis

```bash
# JSON for custom processing
shrike scan . --json > findings.json

# SARIF for GitHub Security tab
shrike scan . --sarif > shrike.sarif
```

## GitHub Action Integration

Add to `.github/workflows/security.yml`:

```yaml
name: Shrike Security Scan

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    permissions:
      security-events: write
    steps:
      - uses: actions/checkout@v4
      - uses: chelate/shrike@v1
```

Results appear in your repository's **Security** tab.

## What Shrike Detects

### Critical: AI Agent Risks

```yaml
# ❌ CRITICAL - Any GitHub user can trigger this
on: issues
jobs:
  ai:
    permissions:
      contents: write
    steps:
      - uses: anthropics/claude-code-action@v1
        with:
          prompt: "${{ github.event.issue.body }}"  # Untrusted!
          claude_args: "--tools Bash"
          allowed_non_write_users: "*"  # No gate!
```

### Critical: Classic Workflow Abuse

```yaml
# ❌ CRITICAL - Pwn-request pattern
on: pull_request_target
jobs:
  test:
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}  # Fork code!
      - run: npm test  # Executes attacker's code
```

## Next Steps

### View Detailed Findings

```bash
# Verbose output shows prerequisites and attack details
shrike scan . --verbose
```

### Scan Remote Repositories

```bash
# Need a GitHub token
export GITHUB_TOKEN=ghp_xxxxx

# Single repo
shrike scan --repo owner/repo-name

# Entire organization
shrike scan --github-org mycompany
```

### Adjust Sensitivity

```bash
# Only fail on critical findings
shrike scan . --fail-on critical

# Show all findings but don't fail
shrike scan . --fail-on low
```

## Common Scenarios

### "I got a CRITICAL finding, now what?"

1. **Read the full details**:
   ```bash
   shrike scan . --verbose
   ```

2. **Check the location**: Navigate to the workflow, job, and step

3. **Understand the prerequisites**: What makes this exploitable?

4. **Fix options**:
   - Add authorization gate: `allowed_non_write_users: ""`
   - Remove untrusted input from prompt
   - Reduce agent permissions
   - Switch to read-only trigger (e.g., `pull_request` instead of `pull_request_target`)

### "I want to scan before every commit"

Add a pre-commit hook:

```bash
# .git/hooks/pre-commit
#!/bin/bash
./shrike scan . --fail-on critical
```

### "I want to integrate with my security dashboard"

Use JSON output:

```bash
shrike scan . --json | \
  jq '.findings[] | {severity, title, location: .Location.WorkflowPath}'
```

## Examples

See the `examples/` directory for:
- **safe-agent-workflow.yml**: Best practices
- **vulnerable-agent-workflow.yml**: What NOT to do
- **cli-examples.sh**: Common CLI patterns
- **github-action-example.yml**: CI/CD integration

## Getting Help

- **Documentation**: See [README.md](README.md)
- **Issues**: https://github.com/chelate/shrike/issues
- **Contributing**: See [CONTRIBUTING.md](CONTRIBUTING.md)

## Quick Reference

```bash
# Local scan
shrike scan .

# Remote scan
shrike scan --repo owner/name

# Org scan
shrike scan --github-org org

# Verbose
shrike scan . -v

# Export JSON
shrike scan . --json

# Export SARIF
shrike scan . --sarif

# Custom report
shrike scan . --report findings.md

# Adjust failure threshold
shrike scan . --fail-on critical
```

---

**Ready to secure your workflows?**

```bash
shrike scan .
```

That's it. Shrike will tell you if there's a risk, who can trigger it, and what they can reach.
