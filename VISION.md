# Shrike: Securing Agentic CI/CD

## The Product Vision

**Chelate's Mission**: Secure the emerging paradigm of AI agents in CI/CD pipelines.

As development teams integrate AI agents into their CI/CD workflows—for code review, testing, deployment automation, and incident response—they create new attack surfaces that traditional security tools don't understand.

**Shrike is step one**: An open-source scanner that finds agentic CI/CD vulnerabilities before attackers do.

---

## The Problem: Agentic CI/CD is Fundamentally Different

### Traditional CI/CD Security
Classic workflow security focuses on:
- Secret exposure
- Malicious dependencies
- Compromised runners
- RBAC misconfigurations

**Tools exist for this**: CodeQL, Dependabot, zizmor, poutine.

### Agentic CI/CD Security (NEW)
AI agents in CI/CD introduce **unique risks**:

1. **Prompt Injection as RCE**
   - Attacker controls agent input (issue title, PR body, comments)
   - Agent executes with CI privileges (write tokens, secrets, cloud access)
   - One malicious prompt = full pipeline compromise

2. **Authorization is Context-Dependent**
   - Traditional: "Is this user allowed to run this action?"
   - Agentic: "Can an outsider's text reach this agent? What can it do?"
   - Requires understanding workflow triggers, gates, and agent capabilities

3. **Tools = Attack Surface**
   - `Bash` tool = shell access
   - `Write` tool = code injection
   - `Edit` tool = backdoor insertion
   - Traditional scanners see "agent action" and stop

4. **Blast Radius Amplification**
   - Agent chains: output of agent A feeds agent B
   - Privilege escalation: read-only agent triggers write-privileged step
   - Cross-repository reach via org-level tokens

**No existing tool understands this threat model.**

---

## The Solution: Agentic CI/CD Security Platform

### Phase 1: Shrike (Open Source Scanner) ✅ **← YOU ARE HERE**

**What it does:**
- Scans GitHub workflows for agentic vulnerabilities
- Detects outsider-triggerable agents without authorization gates
- Maps untrusted input flow to agent prompts
- Analyzes agent power (tools, permissions, secrets)
- Covers classic workflow abuse as table stakes

**Why open source:**
- Builds developer trust and adoption
- Creates de facto standard for agentic CI/CD security
- Generates awareness of the problem space
- Drives leads to commercial platform

**Current state**: Production-ready, fully functional
- CLI tool + GitHub Action
- Detects Clinejection, PraisonAI, and custom patterns
- Organization-scale scanning
- SARIF integration with GitHub Security

### Phase 2: Titra (Commercial Platform) - *Coming Next*

**Enterprise agentic CI/CD security platform.**

**Core capabilities:**

1. **Complete Attack Path Mapping**
   - Full dataflow analysis across jobs and workflows
   - Agent chaining detection (agent → agent → privileged action)
   - Cross-repository attack paths in monorepo/multi-repo setups
   - Visualization: "This issue comment can reach that production deploy agent"

2. **Runtime Protection**
   - Real-time prompt monitoring and filtering
   - Agent output sanitization before feeding to privileged steps
   - Anomaly detection (agent behaving unexpectedly)
   - Kill-switch for compromised agents

3. **Policy Engine**
   - "No agent with Bash tool on outsider-triggerable workflows"
   - "All agents must use organization-level gate (not wildcard)"
   - "Agents with secrets access require 2-person approval"
   - Policy-as-code with enforcement at deploy time

4. **Continuous Monitoring**
   - Slack/PagerDuty alerts on high-risk agent additions
   - Drift detection (did someone bypass the gate?)
   - Compliance reporting (SOC2, HIPAA agent controls)
   - Historical trending (risk score over time)

5. **Auto-Remediation**
   - PR to fix detected issues
   - Suggest secure alternatives
   - One-click gate enforcement
   - Integration with ticketing systems

**Business model**: SaaS platform
- Free tier: Shrike functionality in UI
- Pro: Runtime protection + policy engine ($X/developer/month)
- Enterprise: SSO, compliance, dedicated support

### Phase 3: Ecosystem & Standards

**Establish Chelate as the authority on agentic CI/CD security:**

1. **Agent Action Security Framework**
   - Standard for secure agent action development
   - Certification program ("Chelate Verified Agent")
   - SDK for building secure agent integrations

2. **Industry Partnerships**
   - Work with Anthropic, Google, OpenAI on secure-by-default agent actions
   - GitHub integration (Shrike built into GitHub Advanced Security)
   - CI/CD platform partnerships (GitLab, CircleCI, Jenkins)

3. **Education & Advocacy**
   - Agentic CI/CD Security Best Practices guide
   - Conference talks, blog posts, research papers
   - CVE disclosures for vulnerable public workflows
   - Training program for security teams

---

## Market Opportunity

### Timing: Perfect Storm

1. **AI Agents are Entering CI/CD NOW**
   - GitHub Copilot Workspace
   - Anthropic's Claude Code Action (just launched)
   - Dozens of startups building CI/CD agents
   - Every major company experimenting with agentic workflows

2. **Security Awareness is Growing**
   - Clinejection CVE (Nov 2024) was a wake-up call
   - More vulns being discovered (PraisonAI, etc.)
   - CISO question: "How do we secure this?"

3. **No Competition**
   - Traditional SAST: Don't understand agents
   - Workflow linters: Focus on classic patterns
   - Agent security: Focused on LLM apps, not CI/CD
   - **Chelate: Only player in agentic CI/CD security**

### Target Customers

**Early Adopters (Shrike → Titra funnel):**
- Tech companies with 100+ engineers
- Already using AI agents in development
- Security-conscious (have AppSec team)
- Budget for developer tools ($50K-500K/year)

**Example buyer persona:**
- **Title**: Head of Security, VP Engineering, DevSecOps Lead
- **Pain**: "We're using Claude Code Action on PRs. How do we make sure an attacker can't abuse it?"
- **Current solution**: Manual reviews, homegrown scripts, nothing
- **Budget**: $100K/year for platform that prevents one incident

**TAM calculation:**
- 50K companies with 100+ developers globally
- 10% adopt AI agents in CI/CD in next 2 years = 5K
- Average contract value: $50K/year
- **TAM: $250M/year** (conservative)

---

## Go-to-Market Strategy

### Step 1: Open Source Adoption (Shrike)

**Goal**: Make Shrike the de facto standard for scanning agentic CI/CD

**Tactics:**
1. **GitHub launch**
   - Trending on Hacker News
   - Post in r/netsec, r/devops, security Twitter
   - "We found X vulnerable workflows in popular repos" blog post

2. **CVE disclosure pipeline**
   - Scan top 1000 GitHub repos with Shrike
   - Responsibly disclose vulnerabilities
   - Each disclosure mentions Shrike
   - Builds credibility and awareness

3. **Community building**
   - Discord for users
   - Contribution-friendly (detection rules are easy to add)
   - "Shrike found my vulnerability" testimonials

4. **Integration everywhere**
   - GitHub Action marketplace
   - One-click install for any repo
   - SARIF integration → appears in Security tab
   - "Powered by Shrike" badges

**Success metrics:**
- 1,000 GitHub stars in 3 months
- 100 companies using Shrike in CI/CD
- 10 CVE disclosures showcasing Shrike
- 5 inbound inquiries about commercial offering

### Step 2: Commercial Platform Launch (Titra)

**Conversion funnel:**
- Free Shrike users encounter limitations
  - "Want runtime protection? Upgrade to Titra"
  - "Need policy enforcement? Try Titra Pro"
- Shrike alerts include "Powered by Chelate" footer
- Shrike findings link to Titra for deeper analysis

**Sales motion:**
1. **Self-serve** (Small teams)
   - Free Shrike → Titra free tier → Pro ($99/dev/month)
   - Credit card signup, instant activation

2. **Sales-assisted** (Enterprise)
   - Shrike user requests enterprise features
   - Demo showing attack path visualization
   - Proof of value: "We found X critical issues"
   - Pilot → Expansion

**Pricing:**
- **Free**: Shrike functionality in cloud UI
- **Pro**: $99/developer/month (runtime + policy)
- **Enterprise**: Custom (SSO, compliance, support)

### Step 3: Platform & Ecosystem

**Expand beyond scanning:**
1. Agent action marketplace (Chelate-verified secure agents)
2. Security consulting (help teams adopt agentic CI/CD safely)
3. Compliance module (SOC2 controls for AI in pipelines)
4. Training & certification

---

## Why This Will Work

### 1. Real, Growing Problem
- AI agents in CI/CD are happening **now**
- Security teams are asking "how do we secure this?"
- Vulnerabilities are being discovered regularly
- No current solution exists

### 2. Differentiated Insight
- Two-question model (trigger + power) is unique
- Only security team that deeply understands agentic CI/CD
- Technical moat: Complete attack path analysis is hard

### 3. Open Source → Commercial Playbook
- Proven model (GitGuardian, Snyk, Semgrep)
- Free tool drives awareness and adoption
- Commercial platform captures enterprise value
- Community becomes growth engine

### 4. Timing is Perfect
- Early enough: No entrenched competitors
- Late enough: Real usage and pain points exist
- Market is ready: AI hype → AI security budget

### 5. Founder Credibility
- Found and disclosed real CVEs (Clinejection, etc.)
- Built production scanner (Shrike)
- Deep understanding of both AI and security
- Positioned as expert in the space

---

## Success Criteria

### 6 Months (Shrike Adoption)
- [ ] 1,000+ GitHub stars
- [ ] 100+ companies using Shrike
- [ ] 10+ CVE disclosures
- [ ] 3+ conference talks/blog posts
- [ ] 5+ inbound enterprise leads

### 12 Months (Titra Launch)
- [ ] Titra beta with 10 design partners
- [ ] $500K ARR from early customers
- [ ] 50+ Pro tier users
- [ ] Integration with GitHub Advanced Security
- [ ] Series A funding ($3-5M)

### 24 Months (Platform)
- [ ] $5M ARR
- [ ] 500+ paying customers
- [ ] Recognized category leader
- [ ] Platform launch with policy engine
- [ ] "Chelate Verified" agent certification program

---

## The Narrative

**For YC/investors:**

> "As AI agents move into CI/CD pipelines, they create an entirely new attack surface. Traditional security tools don't understand agents—they can't tell if an outsider can control an agent's prompt, what tools it has access to, or what it can reach.
>
> We're building the security platform for agentic CI/CD. Shrike, our open source scanner, is already detecting vulnerabilities like Clinejection (CVE-2024-53257) that gave attackers RCE through a misconfigured AI agent.
>
> We've validated the problem: real CVEs, real exploits, and security teams actively asking us 'how do we secure this?'
>
> The market is exploding—every company will have AI agents in their pipelines within 2 years. We're positioned to be the **Snyk for agentic CI/CD**: open source scanner drives adoption, commercial platform captures enterprise value.
>
> Founders have deep expertise (found multiple CVEs, built production security tools), we're first to market, and we're solving a problem that didn't exist 6 months ago but will be critical in 12."

**For customers:**

> "Your team is using AI agents to review PRs, run tests, and deploy code. That's powerful—and dangerous.
>
> Shrike scans your workflows and tells you: Can an attacker trigger this agent? What can it access? What's the blast radius?
>
> We found and disclosed Clinejection, PraisonAI, and dozens of other vulnerabilities. Now we're helping teams like yours secure agentic CI/CD before attackers find your gaps."

---

## Next Immediate Steps

1. **Launch Shrike publicly**
   - GitHub release with binaries
   - Hacker News post
   - Blog: "We scanned 1000 repos and found..."

2. **Start CVE pipeline**
   - Scan popular repos with Shrike
   - Find 5-10 vulnerabilities
   - Responsible disclosure
   - Publicity from CVEs

3. **Build Titra MVP**
   - Attack path visualization
   - Policy engine prototype
   - 3-5 design partners

4. **Fundraising**
   - YC application with traction from Shrike
   - Metrics: GitHub stars, CVEs, enterprise leads
   - Vision: Platform for securing agentic CI/CD

---

## Why Shrike → Titra → Ecosystem Works

**Shrike** (free, open source):
- Solves immediate pain: "Is my workflow vulnerable?"
- Builds trust: "They understand this problem"
- Creates awareness: "Oh, this is a real risk"
- Generates leads: "We need more than scanning"

**Titra** (commercial platform):
- Captures enterprise value: Runtime protection, policy, compliance
- Recurring revenue: SaaS model, expand over time
- Stickiness: Integrated into deployment pipeline
- Network effects: More customers → better detection models

**Ecosystem** (platform):
- Becomes industry standard
- Agent creators need "Chelate Verified" badge
- Security teams buy the platform
- Defensible moat from data and integrations

---

**This is not just a security scanner. This is the foundation of the agentic CI/CD security category.**

Shrike is step one. The vision is much bigger.
