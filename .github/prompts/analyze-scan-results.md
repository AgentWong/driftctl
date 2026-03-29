---
name: analyze-scan-results
description: Analyze driftctl scan output from JSON report
---

# Analyze Scan Results

Read the scan results from `test-output/report.json` and provide a summary:

1. **Overall status:** in-sync or drifted
2. **Coverage:** percentage of resources managed by IaC
3. **Resource counts:** managed, unmanaged, deleted, and drifted
4. **Drifted resources:** for each, show the resource type, ID, and what attributes changed
5. **Unmanaged resources:** list by resource type with counts
6. **Alerts:** any warnings from the scan

If `test-output/report.html` exists, mention it can be opened in a browser for a visual report.
