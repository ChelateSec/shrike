#!/bin/bash
# Shrike CLI Usage Examples

# 1. Scan current directory
shrike scan .

# 2. Scan a specific path
shrike scan /path/to/repo

# 3. Scan with verbose output
shrike scan . --verbose

# 4. Export findings as JSON
shrike scan . --json > findings.json

# 5. Export findings as SARIF for GitHub Code Scanning
shrike scan . --sarif > shrike.sarif

# 6. Write detailed markdown report to custom location
shrike scan . --report security-audit.md

# 7. Change failure threshold (don't fail on medium/low)
shrike scan . --fail-on critical

# 8. Scan a remote GitHub repository
export GITHUB_TOKEN=ghp_xxxxx
shrike scan --repo owner/repo-name

# 9. Scan an entire GitHub organization
export GITHUB_TOKEN=ghp_xxxxx
shrike scan --github-org myorg

# 10. Scan org with high concurrency (faster for large orgs)
export GITHUB_TOKEN=ghp_xxxxx
shrike scan --github-org myorg --concurrency 20

# 11. Scan org including archived repositories
export GITHUB_TOKEN=ghp_xxxxx
shrike scan --github-org myorg --include-archived

# 12. Combine options for detailed org scan
export GITHUB_TOKEN=ghp_xxxxx
shrike scan --github-org myorg \
  --concurrency 10 \
  --verbose \
  --report org-security-findings.md \
  --fail-on high

# 13. CI/CD Integration - scan and upload SARIF
shrike scan . --sarif > shrike.sarif
# Then use github/codeql-action/upload-sarif to upload

# 14. Quick check without failing build
shrike scan . --fail-on critical || true

# 15. JSON output for further processing
shrike scan . --json | jq '.findings[] | select(.Severity == "critical")'
