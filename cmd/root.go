package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:   "shrike",
	Short: "Security scanner for AI agents in CI/CD",
	Long: `Shrike - Open-source security scanner for agentic CI/CD

Detects vulnerabilities in GitHub workflows where AI agents can be exploited.

Answers two critical questions:
  1. Can an outsider trigger this agent?
  2. What can it reach?

Website: https://chelate.tech
GitHub:  https://github.com/w3shinew/shrike`,
	Version: version,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(scanCmd)
	rootCmd.SetVersionTemplate(fmt.Sprintf("Shrike v%s\nPart of Chelate Security - https://chelate.tech\n", version))
}
