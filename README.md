# Shrike

**Open-source security scanner for AI agents in CI/CD pipelines.**

<img width="1024" height="304" alt="image" src="https://github.com/user-attachments/assets/c1f84809-9502-441c-96da-5318d28cc37d" />


Shrike detects vulnerabilities in GitHub workflows where AI agents can be exploited—before attackers find them.

---

## What is Shrike?

As AI agents move into CI/CD workflows (code review, testing, deployment), they create new security risks that traditional tools don't catch.

**Shrike answers two critical questions:**
1. **Can an outsider trigger this agent?** (Authorization & triggers)
2. **What can it reach?** (Untrusted input + tools + permissions)

Secures Agentic Workflows.

---

## Quick Start

### Install

**One-line install:**
```bash
curl -fsSL https://chelate.tech/install.sh | bash
```

**Or install manually:**
```bash
# Download the latest release
wget https://github.com/ChelateSec/shrike/releases/latest/download/shrike-linux-amd64 -O shrike
chmod +x shrike
sudo mv shrike /usr/local/bin/
```

---

## Usage

```bash
# Scan current directory
shrike scan .

# Scan a specific path
shrike scan /path/to/repo

# Scan remote GitHub repository (public repos work without token)
shrike scan --repo owner/repo-name

# For private repos or to avoid rate limits, use a token
export GITHUB_TOKEN=ghp_xxxxx
shrike scan --repo owner/private-repo

# Scan entire GitHub organization (requires token)
export GITHUB_TOKEN=ghp_xxxxx
shrike scan --github-org your-org

# Export findings as JSON
shrike scan . --json

# Export findings as SARIF (for GitHub Security tab)
shrike scan . --sarif

# Verbose output with details
shrike scan . --verbose

# Get help
shrike --help
shrike scan --help
```

---

## What It Detects

### Critical: AI Agent Vulnerabilities
- Outsider-triggerable agents without authorization gates
- Untrusted input (issue body, PR title) reaching agent prompts
- Over-privileged agents with dangerous tools (Bash, Write, Edit)
- Agents with access to secrets and write permissions

### Classic Workflow Risks
- Pwn-requests (`pull_request_target` + fork code execution)
- Expression injection (untrusted `${{ }}` in shell commands)
- Over-privileged workflows

---

## Example

**Vulnerable workflow:**
```yaml
on: issues
jobs:
  ai-assist:
    permissions:
      contents: write
    steps:
      - uses: anthropics/claude-code-action@v1
        with:
          prompt: "${{ github.event.issue.body }}"  #  Attacker controls this
          claude_args: "--tools Bash"               #  Shell access
          allowed_non_write_users: "*"              #  No gate!
```

**Shrike output:**
```
[1] CRITICAL - AI agent triggerable by any GitHub user without authorization
    Triggerable: Any GitHub user via issues, no authorization gate
    Impact: Agent execution with contents:write, tools: Bash;
            receives untrusted input: github.event.issue.body
    Location: .github/workflows/ai.yml → ai-assist → step 0
```

---

## GitHub Action

Use Shrike in your CI/CD:

```yaml
name: Security Scan
on: [push, pull_request]

jobs:
  shrike:
    runs-on: ubuntu-latest
    permissions:
      security-events: write
    steps:
      - uses: actions/checkout@v4
      - uses: ChelateSec/shrike@v1
```

Results appear in your repository's **Security** tab.

---

## Why Shrike?

**Traditional security tools don't understand AI agents:**
- - They see "agent action" and stop
- - Don't check authorization gates like `allowed_non_write_users`
- - Don't understand agent tool configurations
- - Miss prompt injection vectors in CI/CD

**Shrike is the first scanner built for agentic CI/CD security.**

---

## Open Source

Shrike is MIT licensed and developed in the open.

**Why open source?**
- Build trust through transparency
- Community-driven detection rules
- Benefit the entire ecosystem
- Establish security standards for AI in CI/CD

---

## Part of Chelate Security

Shrike is the first open-source tool from [**Chelate Security**](https://chelate.tech), focused on securing agentic CI/CD.

We're building the security platform for AI agents in development pipelines—runtime protection, policy enforcement, and complete attack path analysis.

**Want to secure agentic CI/CD in your organization?**
Reach out: security@chelate.tech

---

## Documentation

- **Quick Start**: See usage examples above
- **Contributing**: [CONTRIBUTING.md](CONTRIBUTING.md)
- **Vision**: [VISION.md](VISION.md) - Full product roadmap
- **Examples**: [examples/](examples/) - Safe and vulnerable workflow patterns

---

## Supported Platforms

- - Linux (amd64, arm64)
- - macOS (Intel, Apple Silicon)
- - Windows (amd64)

---

## Credits

Built by the [Chelate Security](https://chelate.tech) team.

Mitigating by real-world vulnerabilities: Securing the Agentic Workflows

---

## License

MIT License - see [LICENSE](LICENSE)

---

**Find AI-agent-in-CI risk before attackers do.** 

**Website**: https://chelate.tech
**GitHub**: https://github.com/ChelateSec/shrike
