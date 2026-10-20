package detector

import (
	"fmt"
	"strings"
)

// Rule interface for detection rules
type Rule interface {
	Evaluate(ctx *EvalContext) *Finding
	ID() string
}

func allRules() []Rule {
	return []Rule{
		&OutsiderTriggerableAgentRule{},
		&UntrustedInputToAgentRule{},
		&OverPrivilegedAgentRule{},
		&PwnRequestRule{},
		&ExpressionInjectionRule{},
	}
}

// Rule 1 & 2: Outsider-triggerable AI agent (combines Q1 trigger + gate check)
type OutsiderTriggerableAgentRule struct{}

func (r *OutsiderTriggerableAgentRule) ID() string {
	return "agent-outsider-triggerable"
}

func (r *OutsiderTriggerableAgentRule) Evaluate(ctx *EvalContext) *Finding {
	// Only applies to agent steps
	if !ctx.Step.IsAgentStep {
		return nil
	}

	// Q1: Is it triggerable by an outsider?
	if !ctx.IsOutsiderTriggerable() {
		return nil // Not outsider-triggerable, skip
	}

	// Q1: Is there a gate?
	hasGate := ctx.HasAuthGate()

	// Q2a: Does untrusted input reach the agent?
	untrustedInputs := ctx.GetUntrustedInputs()
	hasUntrustedInput := len(untrustedInputs) > 0

	// Q2b: Does the agent have power?
	hasPermissions := ctx.HasDangerousPermissions()
	hasSecrets := ctx.HasDangerousSecrets()
	tools := ctx.GetAgentTools()
	hasDangerousTools := len(tools) > 0

	agentPower := hasPermissions || hasSecrets || hasDangerousTools

	// Determine severity based on which legs are present
	var severity string
	var missingLegs []string

	if !hasGate {
		// Missing gate = anyone can trigger
		if hasUntrustedInput && agentPower {
			severity = "critical"
		} else {
			severity = "high"
			if !hasUntrustedInput {
				missingLegs = append(missingLegs, "untrusted input to agent")
			}
			if !agentPower {
				missingLegs = append(missingLegs, "agent power (tools/permissions/secrets)")
			}
		}
	} else {
		// Has gate but still outsider-triggerable (broader gate)
		if hasUntrustedInput && agentPower {
			severity = "medium"
		} else {
			severity = "low"
			if !hasUntrustedInput {
				missingLegs = append(missingLegs, "untrusted input to agent")
			}
			if !agentPower {
				missingLegs = append(missingLegs, "agent power")
			}
		}
	}

	// Build finding
	var title string
	if !hasGate {
		title = "AI agent triggerable by any GitHub user without authorization"
	} else {
		title = "AI agent triggerable by outsiders with limited gate"
	}

	var triggerDesc string
	if !hasGate {
		triggerDesc = fmt.Sprintf("Any GitHub user via %s, no authorization gate", ctx.GetTriggerDescription())
	} else {
		triggerDesc = fmt.Sprintf("External users via %s, limited authorization", ctx.GetTriggerDescription())
	}

	var impactParts []string
	if hasPermissions {
		impactParts = append(impactParts, fmt.Sprintf("runs with %s", ctx.GetPermissionsDescription()))
	}
	if hasDangerousTools {
		impactParts = append(impactParts, fmt.Sprintf("has tools: %s", strings.Join(tools, ", ")))
	}
	if hasSecrets {
		impactParts = append(impactParts, "has access to secrets")
	}
	if hasUntrustedInput {
		impactParts = append(impactParts, fmt.Sprintf("receives untrusted input: %s", strings.Join(untrustedInputs, ", ")))
	}

	impact := "Agent execution"
	if len(impactParts) > 0 {
		impact += ": " + strings.Join(impactParts, "; ")
	}

	if len(missingLegs) > 0 {
		impact += fmt.Sprintf(". Missing: %s", strings.Join(missingLegs, ", "))
	}

	// Build prerequisites
	var prereqs []string
	if !hasGate {
		prereqs = append(prereqs, "No authorization gate (allowed_non_write_users: * or missing author check)")
	}
	if hasUntrustedInput {
		prereqs = append(prereqs, fmt.Sprintf("Prompt interpolates %s", strings.Join(untrustedInputs, ", ")))
	}
	if hasDangerousTools {
		prereqs = append(prereqs, fmt.Sprintf("Agent has %s enabled", strings.Join(tools, ", ")))
	}

	agentName := "AI agent"
	if ctx.Step.AgentInfo != nil {
		agentName = ctx.Step.AgentInfo.Name
	}

	return &Finding{
		ID:             r.ID(),
		Severity:       severity,
		Title:          title,
		Description:    fmt.Sprintf("%s step on outsider-triggerable workflow", agentName),
		Triggerability: triggerDesc,
		Impact:         impact,
		Location:       ctx.Location(),
		Prerequisites:  prereqs,
		RuleIDs:        []string{r.ID()},
	}
}

// Rule 3: Untrusted input to agent (Q2a focused)
type UntrustedInputToAgentRule struct{}

func (r *UntrustedInputToAgentRule) ID() string {
	return "agent-untrusted-input"
}

func (r *UntrustedInputToAgentRule) Evaluate(ctx *EvalContext) *Finding {
	// Only for agent steps
	if !ctx.Step.IsAgentStep {
		return nil
	}

	untrustedInputs := ctx.GetUntrustedInputs()
	if len(untrustedInputs) == 0 {
		return nil
	}

	// This is covered by OutsiderTriggerableAgentRule if outsider-triggerable
	// Only report separately if not outsider-triggerable (informational)
	if ctx.IsOutsiderTriggerable() {
		return nil // Already covered by main rule
	}

	// Informational: trusted trigger but still interpolating untrusted data
	return &Finding{
		ID:             r.ID(),
		Severity:       "low",
		Title:          "Agent prompt interpolates untrusted input (trusted trigger)",
		Description:    "AI agent receives potentially untrusted input, but requires trusted trigger",
		Triggerability: fmt.Sprintf("Requires trusted trigger: %s", ctx.GetTriggerDescription()),
		Impact:         fmt.Sprintf("Untrusted input reaches agent: %s", strings.Join(untrustedInputs, ", ")),
		Location:       ctx.Location(),
		RuleIDs:        []string{r.ID()},
	}
}

// Rule 4: Over-privileged agent (Q2b focused)
type OverPrivilegedAgentRule struct{}

func (r *OverPrivilegedAgentRule) ID() string {
	return "agent-over-privileged"
}

func (r *OverPrivilegedAgentRule) Evaluate(ctx *EvalContext) *Finding {
	// Only for agent steps
	if !ctx.Step.IsAgentStep {
		return nil
	}

	hasPermissions := ctx.HasDangerousPermissions()
	hasSecrets := ctx.HasDangerousSecrets()
	tools := ctx.GetAgentTools()

	// If already covered by outsider-triggerable rule, skip
	if ctx.IsOutsiderTriggerable() {
		return nil
	}

	// Report if agent has power but is not in a risky context
	if !hasPermissions && !hasSecrets {
		return nil
	}

	var impactParts []string
	if hasPermissions {
		impactParts = append(impactParts, ctx.GetPermissionsDescription())
	}
	if hasSecrets {
		impactParts = append(impactParts, "access to secrets")
	}
	if len(tools) > 0 {
		impactParts = append(impactParts, fmt.Sprintf("tools: %s", strings.Join(tools, ", ")))
	}

	return &Finding{
		ID:             r.ID(),
		Severity:       "medium",
		Title:          "AI agent with elevated permissions (trusted trigger)",
		Description:    "Agent has write permissions or secret access, ensure trigger is properly restricted",
		Triggerability: fmt.Sprintf("Requires trusted trigger: %s", ctx.GetTriggerDescription()),
		Impact:         strings.Join(impactParts, ", "),
		Location:       ctx.Location(),
		RuleIDs:        []string{r.ID()},
	}
}

// Classic workflow rule: pwn-request pattern
type PwnRequestRule struct{}

func (r *PwnRequestRule) ID() string {
	return "pwn-request"
}

func (r *PwnRequestRule) Evaluate(ctx *EvalContext) *Finding {
	// Check for pull_request_target or workflow_run
	isPRTarget := false
	for _, trigger := range ctx.Workflow.Triggers {
		if trigger.Type == "pull_request_target" || trigger.Type == "workflow_run" {
			isPRTarget = true
			break
		}
	}

	if !isPRTarget {
		return nil
	}

	// Check if this step checks out fork code
	checkoutsFork := false
	if ctx.Step.Uses != "" && strings.Contains(ctx.Step.Uses, "actions/checkout") {
		// Check ref parameter
		if ref, ok := ctx.Step.With["ref"]; ok {
			refStr := fmt.Sprintf("%v", ref)
			// Look for fork refs
			if strings.Contains(refStr, "github.event.pull_request.head") ||
				strings.Contains(refStr, "refs/pull/") {
				checkoutsFork = true
			}
		}
	}

	if !checkoutsFork {
		return nil
	}

	// Check if code is executed after checkout
	hasExecution := false
	// Look at subsequent steps
	for i := ctx.StepIndex + 1; i < len(ctx.Job.Steps); i++ {
		nextStep := ctx.Job.Steps[i]
		if nextStep.Run != "" || (nextStep.Uses != "" && !strings.Contains(nextStep.Uses, "actions/checkout")) {
			hasExecution = true
			break
		}
	}

	if !hasExecution {
		return nil
	}

	var severity string
	if ctx.HasDangerousPermissions() {
		severity = "critical"
	} else {
		severity = "high"
	}

	return &Finding{
		ID:          r.ID(),
		Severity:    severity,
		Title:       "Pwn-request: pull_request_target with fork code execution",
		Description: "Workflow checks out and executes code from a fork with elevated privileges",
		Triggerability: "Any GitHub user via pull request from fork",
		Impact:      fmt.Sprintf("RCE with %s; can compromise workflow", ctx.GetPermissionsDescription()),
		Location:    ctx.Location(),
		Prerequisites: []string{
			"pull_request_target or workflow_run trigger",
			"Checkout of fork code (github.event.pull_request.head.*)",
			"Execution of checked-out code",
		},
		RuleIDs: []string{r.ID()},
	}
}

// Classic workflow rule: expression injection
type ExpressionInjectionRule struct{}

func (r *ExpressionInjectionRule) ID() string {
	return "expression-injection"
}

func (r *ExpressionInjectionRule) Evaluate(ctx *EvalContext) *Finding {
	// Skip agent steps (covered by other rules)
	if ctx.Step.IsAgentStep {
		return nil
	}

	// Check run commands for direct interpolation of untrusted expressions
	if ctx.Step.Run == "" {
		return nil
	}

	var injectedExprs []string
	for _, expr := range ctx.GetUntrustedInputs() {
		// Check if expression is directly in run command (not in a safe context)
		if strings.Contains(ctx.Step.Run, "${{ "+expr) || strings.Contains(ctx.Step.Run, "${"+expr) {
			injectedExprs = append(injectedExprs, expr)
		}
	}

	if len(injectedExprs) == 0 {
		return nil
	}

	var severity string
	if ctx.IsOutsiderTriggerable() {
		severity = "critical"
	} else {
		severity = "medium"
	}

	var triggerDesc string
	if ctx.IsOutsiderTriggerable() {
		triggerDesc = fmt.Sprintf("Any GitHub user via %s", ctx.GetTriggerDescription())
	} else {
		triggerDesc = fmt.Sprintf("Trusted trigger: %s", ctx.GetTriggerDescription())
	}

	return &Finding{
		ID:          r.ID(),
		Severity:    severity,
		Title:       "Expression injection: untrusted input in run command",
		Description: "Untrusted GitHub context directly interpolated into shell command",
		Triggerability: triggerDesc,
		Impact:      fmt.Sprintf("Shell injection via %s; can execute arbitrary commands", strings.Join(injectedExprs, ", ")),
		Location:    ctx.Location(),
		Prerequisites: []string{
			fmt.Sprintf("Direct interpolation of %s in run command", strings.Join(injectedExprs, ", ")),
		},
		VulnerableLines: []string{ctx.Step.Run},
		RuleIDs:         []string{r.ID()},
	}
}
