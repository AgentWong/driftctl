# Developer guide

This directory contains documentation about the driftctl codebase, aimed at readers who are interested in making code contributions.

- [Architecture overview](#architecture-overview)
- [Adding new resource type mappings](new-resource.md)
- [Testing](testing.md)
- [Middlewares](middlewares.md)

## Architecture overview

driftctl is an AWS-focused Terraform drift detection tool. It supports two scan modes:

### Inventory mode (default)

1. **AWS Config inventory** — A single `ConfigEnumerator` calls the AWS Config API to discover all resources in the account, then maps them to Terraform resource types using a static mapping table (132 mappings).
2. **Terraform state reading** — Terraform state files are read (from local files, S3 backends, etc.) via Terraform providers + gRPC.
3. **Comparison** — Resources from Config are compared against Terraform state to identify unmanaged resources.
4. **Middlewares** — Reconciliation middlewares normalize resources before comparison (e.g. expanding inline routes, reconciling IDs).
5. **Categorization** — Unmanaged resources are categorized (CloudFormation-managed, service-linked, unsupported) to reduce false positives.

### Plan mode (`--mode plan`)

1. **Terraform plan** — Runs `terraform plan` via `terraform-exec` against a specified root module to detect attribute-level drift.
2. **AWS Config inventory** — Same Config-based discovery as inventory mode.
3. **Combined analysis** — Plan results (drifted/deleted resources) are merged with Config inventory (unmanaged detection).
4. **Categorization** — Same categorizer chain as inventory mode.

## Core concepts

- `Remote` represents the AWS cloud provider
- `Resource` is an abstract representation of a cloud resource (e.g. S3 bucket, EC2 instance)
- `BulkEnumerator` discovers all supported resource types in a single API call (used by the Config enumerator)
- `Enumerator` discovers resources of a single type (legacy interface, kept for compatibility)
- `Categorizer` classifies resources into categories for false positive filtering

## Key directories

| Directory | Purpose |
|-----------|---------|
| `enumeration/remote/aws/` | AWS Config enumerator, repository, resource mapping, provider setup |
| `enumeration/remote/common/` | `RemoteLibrary`, `BulkEnumerator`/`Enumerator` interfaces |
| `enumeration/remote/` | Scanner, alerter, error handling |
| `enumeration/resource/aws/` | AWS resource type definitions and metadata |
| `pkg/analyser/` | Analysis model, plan analyzer |
| `pkg/terraform/plan/` | Terraform plan runner and parser |
| `pkg/categorizer/` | Resource categorization framework |
| `pkg/cmd/scan/output/` | Console, JSON, HTML output formatters |
| `pkg/middlewares/` | Resource reconciliation middlewares (AWS-only) |
| `pkg/driftctl.go` | Main orchestrator (`DriftCTL.Run()`) |

## Terminology

- `Remote` is a representation of a cloud provider (AWS only)
- `Resource` is an abstract representation of a cloud provider resource (e.g. S3 bucket, EC2 instance, etc.)
- `BulkEnumerator` lists all supported resource types in one call via AWS Config
- `Enumerator` lists resources of a given type from a given remote (legacy, per-type)
- `Categorizer` classifies a resource into a category (cloudformation_managed, default_resources, service_linked, unsupported, etc.)
