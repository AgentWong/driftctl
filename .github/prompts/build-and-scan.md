---
name: build-and-scan
description: Build driftctl, fix errors, and run a scan against AWS
---

# Build and Scan

Build the driftctl binary and run a scan against AWS. Fix any errors encountered along the way.

## Step 1: Build

```bash
make dev
```

If the build fails, read the compiler errors, fix the source files, and rebuild. Repeat until successful.

## Step 2: Scan

```bash
./bin/driftctl scan \
  --from 'tfstate+s3://terraform-state-07027b6d-e4ba-4f0a-abcf-1520f93ebd4d//**' \
  --output console:// \
  --output json://test-output/report.json \
  --output html://test-output/report.html
```

## Step 3: Handle Results

- **Exit code 0:** in sync, report success
- **Exit code 1:** drift detected, summarize the drift from `test-output/report.json`
- **Exit code 2:** runtime error, diagnose the error, fix the source, and loop back to Step 1

For runtime errors, common causes include:
- AWS credentials not configured or expired (run `aws sso login`)
- AWS Config recorder not enabled in the target region
- Missing or incorrect Terraform provider

Repeat the build-scan loop until the scan completes without runtime errors.

Report the final exit code and summarize any drift or errors found. Mention that `test-output/report.html` can be opened in a browser for a visual report.
