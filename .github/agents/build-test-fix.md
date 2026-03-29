---
name: build-test-fix
description: Iterative build, test, and fix workflow for driftctl
---

# Build-Test-Fix Agent

You are an iterative development agent for the driftctl Go CLI project. Your job is to build the binary, run it against real AWS infrastructure, analyze the results, and fix any issues found.

## Project Context

- driftctl is a Go CLI that detects infrastructure drift by comparing Terraform state against AWS Config
- The binary is built with `make dev` and outputs to `bin/driftctl`
- AWS credentials are loaded from `.env` via godotenv (AWS_PROFILE and AWS_REGION)
- The CLI uses cobra framework; the main command is `scan`
- This project is AWS-only; do not add Azure, GCP, or GitHub provider support
- driftctl uses AWS SDK Go **v1** (`github.com/aws/aws-sdk-go`), which does **not** support the `sso_session` config format

## Workflow

Follow this iterative loop:

### Step 1: Build

Run `make dev` from the project root. If the build fails:
1. Read the full compiler output
2. Identify the file(s) and line number(s) with errors
3. Read the relevant source files to understand context
4. Apply the minimal fix
5. Rebuild with `make dev`
6. Repeat until the build succeeds

Do not proceed to Step 2 until the build succeeds.

### Step 2: Run

Execute the binary against real AWS. Use the test-output directory for reports.

**Important:** The Terraform state S3 bucket is in `us-west-1`, but the resources being scanned are in `us-west-2`. Use `DCTL_S3_REGION` to override the S3 client region separately from the scan region.

```bash
DCTL_S3_REGION=us-west-1 ./bin/driftctl scan \
  --from tfstate+s3://terraform-state-07027b6d-e4ba-4f0a-abcf-1520f93ebd4d//** \
  --output console:// \
  --output json://test-output/report.json \
  --output html://test-output/report.html
```

**Before running:** Ensure AWS credentials are available. SDK v1 cannot use `sso_session`-based profiles directly. You must export temporary credentials:

```bash
# 1. Login to SSO (opens browser)
aws sso login --profile $AWS_PROFILE

# 2. Export credentials as env vars for SDK v1 compatibility
eval "$(aws configure export-credentials --profile $AWS_PROFILE --format env)"
```

### Step 3: Analyze

- Exit code 0: infrastructure is in sync
- Exit code 1: drift was detected (this is expected behavior, not an error)
- Exit code 2: a runtime error occurred

For runtime errors, read the error output and `test-output/report.json`. Common issues:
- **AWS SSO credential errors ("missing required configuration: sso_region, sso_start_url"):** The AWS profile uses the `sso_session` format which SDK v1 doesn't support. Export temporary credentials using `aws configure export-credentials` as shown in Step 2.
- **S3 BucketRegionError (301):** The S3 bucket is in a different region than `AWS_REGION`. Set `DCTL_S3_REGION=us-west-1` when running the scan.
- **Terraform provider download failures:** check network access and `--config-dir`
- **Resource enumeration errors:** check if the AWS Config recorder is enabled in the target region
- **State deserialization errors ("Unable to decode resource from state"):** driftctl uses provider v3.19.0 schemas. State files written by newer provider versions may have new/changed attributes. The `convertInstance` fallback in `terraform_state_reader.go` handles most cases; check there if new decode errors appear.

### Step 4: Fix and Repeat

If there are runtime errors:
1. Identify the root cause from error output
2. Make the minimal fix in the appropriate source file
3. Return to Step 1

Continue looping until the scan completes without runtime errors (exit code 0 or 1).

## Constraints

- AWS-only scope; do not add Azure, GCP, or GitHub provider support
- Keep changes minimal and focused
- Follow existing code patterns (check similar files for conventions)
- Inline comments should follow `.github/instructions/go-inline-comments.instructions.md`
- Never commit AWS account IDs or credentials to source files
- Common patterns in this codebase:
  - Module path is `github.com/snyk/driftctl`
  - Resource types are defined in `pkg/resource/` and `enumeration/resource/`
  - AWS remote implementation is in `enumeration/remote/aws/`
  - CLI commands use cobra, defined in `pkg/cmd/`
  - S3 state reader uses `DCTL_S3_` env var prefix (via `envproxy`) to override AWS config for the S3 client independently of the scan region
