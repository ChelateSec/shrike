package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/chelate/shrike/internal/collector"
	"github.com/chelate/shrike/internal/detector"
	"github.com/chelate/shrike/internal/parser"
	"github.com/chelate/shrike/internal/reporter"
	"github.com/spf13/cobra"
)

var (
	repoFlag      string
	orgFlag       string
	jsonOutput    bool
	sarifOutput   bool
	reportPath    string
	verbose       bool
	failOnLevel   string
	concurrency   int
	includeArchived bool
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan a repository for AI agent and workflow risks",
	Long: `Scan GitHub workflows for security risks:
- AI agents triggerable by outsiders
- Untrusted input reaching agents
- Over-privileged agents and workflows
- Classic workflow abuse patterns (pwn-requests, expression injection, etc.)

Three scan modes:
  shrike scan <path>              - Scan local repository
  shrike scan --repo owner/name   - Scan remote repository via GitHub API
  shrike scan --github-org org    - Scan all repositories in organization`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScan,
}

func init() {
	scanCmd.Flags().StringVar(&repoFlag, "repo", "", "Scan a GitHub repository (format: owner/name)")
	scanCmd.Flags().StringVar(&orgFlag, "github-org", "", "Scan all repositories in a GitHub organization")
	scanCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output findings as JSON to stdout")
	scanCmd.Flags().BoolVar(&sarifOutput, "sarif", false, "Output findings as SARIF to stdout")
	scanCmd.Flags().StringVar(&reportPath, "report", "", "Write detailed findings report to file (default: shrike-findings.md if findings exist)")
	scanCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed finding information")
	scanCmd.Flags().StringVar(&failOnLevel, "fail-on", "high", "Fail (exit non-zero) on this severity or higher (critical, high, medium, low)")
	scanCmd.Flags().IntVar(&concurrency, "concurrency", 5, "Number of concurrent repository scans (org mode only)")
	scanCmd.Flags().BoolVar(&includeArchived, "include-archived", false, "Include archived repositories (org mode only)")
}

func runScan(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Determine scan target
	var scanTarget collector.ScanTarget

	if orgFlag != "" {
		if repoFlag != "" || len(args) > 0 {
			return fmt.Errorf("cannot combine --github-org with --repo or path argument")
		}
		scanTarget = collector.ScanTarget{
			Type:            collector.TargetOrg,
			Org:             orgFlag,
			Concurrency:     concurrency,
			IncludeArchived: includeArchived,
		}
	} else if repoFlag != "" {
		if len(args) > 0 {
			return fmt.Errorf("cannot combine --repo with path argument")
		}
		parts := strings.Split(repoFlag, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repo format, expected owner/name")
		}
		scanTarget = collector.ScanTarget{
			Type:  collector.TargetRepo,
			Owner: parts[0],
			Repo:  parts[1],
		}
	} else {
		// Local path mode
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		scanTarget = collector.ScanTarget{
			Type: collector.TargetLocal,
			Path: path,
		}
	}

	// Get GitHub token for remote scans
	var token string
	if scanTarget.Type != collector.TargetLocal {
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return fmt.Errorf("GITHUB_TOKEN environment variable required for remote scans")
		}
	}

	// Stage 1: Collect workflows
	c := collector.New(token)
	workflows, err := c.Collect(ctx, scanTarget)
	if err != nil {
		return fmt.Errorf("failed to collect workflows: %w", err)
	}

	if len(workflows) == 0 {
		fmt.Println("No workflows found")
		return nil
	}

	// Stage 2: Parse workflows
	p := parser.New()
	parsedWorkflows := make([]*parser.ParsedWorkflow, 0, len(workflows))
	for _, wf := range workflows {
		parsed, err := p.Parse(wf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", wf.Path, err)
			continue
		}
		parsedWorkflows = append(parsedWorkflows, parsed)
	}

	// Stage 3: Identify agent steps
	agentCount := 0
	for _, wf := range parsedWorkflows {
		for _, job := range wf.Jobs {
			for _, step := range job.Steps {
				if step.IsAgentStep {
					agentCount++
				}
			}
		}
	}

	fmt.Printf("Scanned %d workflow(s) across %d repository/repositories\n", len(parsedWorkflows), scanTarget.RepoCount())
	if agentCount > 0 {
		fmt.Printf("Found %d AI agent step(s)\n", agentCount)
	}

	// Stage 4: Evaluate (detect risks)
	d := detector.New()
	findings := d.Evaluate(parsedWorkflows)

	// Stage 5: Report
	r := reporter.New(reporter.Config{
		Verbose:    verbose,
		ReportPath: reportPath,
	})

	if jsonOutput {
		return r.WriteJSON(os.Stdout, findings)
	} else if sarifOutput {
		return r.WriteSARIF(os.Stdout, findings)
	} else {
		r.WriteTerminal(os.Stdout, findings, len(parsedWorkflows), scanTarget.RepoCount())

		// Auto-write report file if there are findings and no explicit report path
		if len(findings) > 0 && reportPath == "" {
			defaultReport := "shrike-findings.md"
			if err := r.WriteMarkdownReport(defaultReport, findings); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to write report to %s: %v\n", defaultReport, err)
			} else {
				fmt.Printf("\nDetailed findings written to %s\n", defaultReport)
			}
		} else if reportPath != "" {
			if err := r.WriteMarkdownReport(reportPath, findings); err != nil {
				return fmt.Errorf("failed to write report: %w", err)
			}
			fmt.Printf("\nDetailed findings written to %s\n", reportPath)
		}
	}

	// Exit with error if findings meet fail-on threshold
	return checkFailureThreshold(findings, failOnLevel)
}

func checkFailureThreshold(findings []*detector.Finding, failOnLevel string) error {
	if len(findings) == 0 {
		return nil
	}

	severityRank := map[string]int{
		"critical": 4,
		"high":     3,
		"medium":   2,
		"low":      1,
	}

	threshold, ok := severityRank[failOnLevel]
	if !ok {
		return fmt.Errorf("invalid fail-on level: %s", failOnLevel)
	}

	for _, f := range findings {
		if severityRank[f.Severity] >= threshold {
			return fmt.Errorf("found %s severity finding, failing as requested", f.Severity)
		}
	}

	return nil
}
