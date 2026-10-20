package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/chelate/shrike/internal/detector"
)

type Config struct {
	Verbose    bool
	ReportPath string
}

type Reporter struct {
	config Config
}

func New(config Config) *Reporter {
	return &Reporter{config: config}
}

// WriteTerminal writes findings to terminal (stdout) - the default clean view
func (r *Reporter) WriteTerminal(w io.Writer, findings []*detector.Finding, workflowCount, repoCount int) {
	if len(findings) == 0 {
		repos := "repository"
		if repoCount > 1 {
			repos = fmt.Sprintf("%d repositories", repoCount)
		}
		fmt.Fprintf(w, "\nNo triggerable agent or workflow risks found in %d workflows across %s\n", workflowCount, repos)
		return
	}

	fmt.Fprintf(w, "\n")
	for i, f := range findings {
		r.writeTerminalFinding(w, f, i+1)
		fmt.Fprintf(w, "\n")
	}

	fmt.Fprintf(w, "Found %d finding(s)\n", len(findings))
}

func (r *Reporter) writeTerminalFinding(w io.Writer, f *detector.Finding, num int) {
	// Color codes
	var severityColor string
	switch f.Severity {
	case "critical":
		severityColor = "\033[1;31m" // Bold red
	case "high":
		severityColor = "\033[0;31m" // Red
	case "medium":
		severityColor = "\033[0;33m" // Yellow
	case "low":
		severityColor = "\033[0;36m" // Cyan
	default:
		severityColor = "\033[0m"
	}
	reset := "\033[0m"

	// Check if stdout is a terminal (for color support)
	if !isTerminal(w) {
		severityColor = ""
		reset = ""
	}

	fmt.Fprintf(w, "%s[%d] %s%s - %s\n", severityColor, num, strings.ToUpper(f.Severity), reset, f.Title)
	fmt.Fprintf(w, "    Triggerable: %s\n", f.Triggerability)
	fmt.Fprintf(w, "    Impact: %s\n", f.Impact)
	fmt.Fprintf(w, "    Location: %s\n", f.Location.String())

	if r.config.Verbose && len(f.Prerequisites) > 0 {
		fmt.Fprintf(w, "    Prerequisites:\n")
		for _, p := range f.Prerequisites {
			fmt.Fprintf(w, "      - %s\n", p)
		}
	}
}

// WriteMarkdownReport writes detailed findings to a markdown file
func (r *Reporter) WriteMarkdownReport(path string, findings []*detector.Finding) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "# Shrike Security Findings\n\n")
	fmt.Fprintf(f, "Generated: %s\n\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(f, "Total findings: %d\n\n", len(findings))

	if len(findings) == 0 {
		fmt.Fprintf(f, "No security risks detected.\n")
		return nil
	}

	// Summary by severity
	severityCounts := make(map[string]int)
	for _, finding := range findings {
		severityCounts[finding.Severity]++
	}

	fmt.Fprintf(f, "## Summary\n\n")
	for _, severity := range []string{"critical", "high", "medium", "low"} {
		if count := severityCounts[severity]; count > 0 {
			fmt.Fprintf(f, "- **%s**: %d\n", strings.ToUpper(severity), count)
		}
	}
	fmt.Fprintf(f, "\n---\n\n")

	// Detailed findings
	for i, finding := range findings {
		r.writeMarkdownFinding(f, finding, i+1)
		fmt.Fprintf(f, "\n---\n\n")
	}

	return nil
}

func (r *Reporter) writeMarkdownFinding(w io.Writer, f *detector.Finding, num int) {
	fmt.Fprintf(w, "## Finding %d: %s\n\n", num, f.Title)

	fmt.Fprintf(w, "**Severity**: %s  \n", strings.ToUpper(f.Severity))
	fmt.Fprintf(w, "**Vulnerability**: %s\n\n", f.Description)

	fmt.Fprintf(w, "### Location\n\n")
	fmt.Fprintf(w, "- **Workflow**: `%s`\n", f.Location.WorkflowPath)
	if f.Location.WorkflowName != "" {
		fmt.Fprintf(w, "- **Workflow name**: %s\n", f.Location.WorkflowName)
	}
	fmt.Fprintf(w, "- **Job**: `%s`", f.Location.JobID)
	if f.Location.JobName != "" {
		fmt.Fprintf(w, " (%s)", f.Location.JobName)
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "- **Step**: %d", f.Location.StepIndex)
	if f.Location.StepName != "" {
		fmt.Fprintf(w, " (`%s`)", f.Location.StepName)
	}
	fmt.Fprintf(w, "\n\n")

	if len(f.VulnerableLines) > 0 {
		fmt.Fprintf(w, "### Vulnerable Lines\n\n")
		fmt.Fprintf(w, "```yaml\n")
		for _, line := range f.VulnerableLines {
			fmt.Fprintf(w, "%s\n", line)
		}
		fmt.Fprintf(w, "```\n\n")
	}

	fmt.Fprintf(w, "### Can an outsider trigger it?\n\n")
	fmt.Fprintf(w, "%s\n\n", f.Triggerability)

	if len(f.Prerequisites) > 0 {
		fmt.Fprintf(w, "### Prerequisites\n\n")
		for _, p := range f.Prerequisites {
			fmt.Fprintf(w, "- %s\n", p)
		}
		fmt.Fprintf(w, "\n")
	}

	if f.Repro != "" {
		fmt.Fprintf(w, "### Reproduction\n\n")
		fmt.Fprintf(w, "%s\n\n", f.Repro)
	}

	fmt.Fprintf(w, "### Impact\n\n")
	fmt.Fprintf(w, "%s\n", f.Impact)
}

// WriteJSON writes findings as JSON
func (r *Reporter) WriteJSON(w io.Writer, findings []*detector.Finding) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(map[string]interface{}{
		"findings": findings,
		"count":    len(findings),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// WriteSARIF writes findings in SARIF format for GitHub Code Scanning
func (r *Reporter) WriteSARIF(w io.Writer, findings []*detector.Finding) error {
	sarif := r.buildSARIF(findings)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(sarif)
}

func (r *Reporter) buildSARIF(findings []*detector.Finding) map[string]interface{} {
	// Build SARIF rules from findings
	rulesMap := make(map[string]map[string]interface{})

	for _, f := range findings {
		for _, ruleID := range f.RuleIDs {
			if _, exists := rulesMap[ruleID]; !exists {
				rulesMap[ruleID] = map[string]interface{}{
					"id":   ruleID,
					"name": ruleID,
					"shortDescription": map[string]interface{}{
						"text": f.Title,
					},
					"fullDescription": map[string]interface{}{
						"text": f.Description,
					},
					"help": map[string]interface{}{
						"text": f.Impact,
					},
					"defaultConfiguration": map[string]interface{}{
						"level": r.severityToSARIFLevel(f.Severity),
					},
					"properties": map[string]interface{}{
						"tags": []string{"security", "ai-agents", "workflows"},
					},
				}
			}
		}
	}

	rules := make([]interface{}, 0, len(rulesMap))
	for _, rule := range rulesMap {
		rules = append(rules, rule)
	}

	// Build SARIF results
	results := make([]interface{}, 0, len(findings))
	for _, f := range findings {
		ruleID := f.ID
		if len(f.RuleIDs) > 0 {
			ruleID = f.RuleIDs[0]
		}

		result := map[string]interface{}{
			"ruleId": ruleID,
			"level":  r.severityToSARIFLevel(f.Severity),
			"message": map[string]interface{}{
				"text": fmt.Sprintf("%s\n\nTriggerable: %s\nImpact: %s", f.Description, f.Triggerability, f.Impact),
			},
			"locations": []interface{}{
				map[string]interface{}{
					"physicalLocation": map[string]interface{}{
						"artifactLocation": map[string]interface{}{
							"uri": f.Location.WorkflowPath,
						},
						"region": map[string]interface{}{
							"startLine": f.Location.StepIndex + 1, // Approximate
						},
					},
				},
			},
		}

		results = append(results, result)
	}

	return map[string]interface{}{
		"version": "2.1.0",
		"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		"runs": []interface{}{
			map[string]interface{}{
				"tool": map[string]interface{}{
					"driver": map[string]interface{}{
						"name":            "Shrike",
						"version":         "1.0.0",
						"informationUri":  "https://github.com/chelate/shrike",
						"semanticVersion": "1.0.0",
						"rules":           rules,
					},
				},
				"results": results,
			},
		},
	}
}

func (r *Reporter) severityToSARIFLevel(severity string) string {
	switch severity {
	case "critical", "high":
		return "error"
	case "medium":
		return "warning"
	case "low":
		return "note"
	default:
		return "warning"
	}
}

func isTerminal(w io.Writer) bool {
	// Check if output is to a terminal
	if f, ok := w.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return false
		}
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
	return false
}
