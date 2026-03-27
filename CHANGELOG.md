# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-03-27

### Added
- **AWS Config-based inventory system** — replaced 103 individual AWS resource enumerators with a single AWS Config API integration (`config_repository.go`, `config_enumerator.go`, `config_resource_mapping.go`) covering 80+ resource type mappings
- **BulkEnumerator interface** — new interface in `enumeration/remote/common/library.go` enabling single-call enumeration of all supported resource types, coexisting with the legacy Enumerator interface
- **Terraform plan-based drift detection** — new `pkg/terraform/plan/` package with `runner.go` (terraform-exec integration) and `parser.go` (terraform-json plan parsing) for true configuration drift analysis
- **Plan analyzer** — `pkg/analyser/plan_analyzer.go` combining Config inventory with terraform plan results to classify resources as drifted, deleted, or unmanaged
- **`--mode=plan` scan mode** — new scan mode using terraform plan for drift detection (default `--mode=inventory` preserves backward compatibility)
- **`--terraform-dir` flag** — specify terraform working directory for plan-based scanning
- **Resource categorization framework** — new `pkg/categorizer/` package with pluggable categorizer chain:
  - `cloudformation.go` — detects CloudFormation-managed resources via `aws:cloudformation:stack-name` tag
  - `service_linked.go` — identifies AWS service-linked roles via path/name pattern matching
  - `unsupported.go` — flags resources not covered by Config API mapping
- **`--exclude-category` flag** — filter scan results by category (cloudformation_managed, service_linked, unsupported)
- **Drift detail rendering** — console, JSON, and HTML output formatters updated to display attribute-level drift changes and category groupings

### Changed
- **Scanner** (`enumeration/remote/scanner.go`) — updated to handle both BulkEnumerator (called first) and individual Enumerator types
- **`aws/init.go`** — rewritten from 103 individual `AddEnumerator()` calls to a single `AddBulkEnumerator()` registration
- **Analysis model** (`pkg/analyser/analysis.go`) — extended with `drifted` resources field, `TotalDrifted` in summary, updated `IsSync()` logic
- **`pkg/driftctl.go`** — added plan-based flow with mode selection; removed non-AWS middleware registrations
- **`pkg/cmd/scan.go`** — added `--terraform-dir`, `--mode`, and `--exclude-category` flags; removed Azure CLI flags and non-AWS remote validation

### Removed
- **Azure provider** (~90 files) — all `azurerm` enumerators, repositories, resource types, middlewares, state backend readers, and test schemas
- **Google Cloud provider** (~100 files) — all `google` enumerators, repositories, resource types, middlewares, and test schemas
- **GitHub provider** (~25 files) — all `github` enumerators, repositories, resource types, and test schemas
- **103 individual AWS enumerators** — replaced by the Config-based BulkEnumerator; all `*_enumerator.go` files removed (except `config_enumerator.go`)
- **23 individual AWS repository files** — replaced by `config_repository.go`; old `ec2_repository.go`, `s3_repository.go`, `iam_repository.go`, etc. removed
- **Associated mock and test files** — all `mock_*.go` and `*_test.go` files for deleted enumerators and repositories
- **Non-AWS provider constants and references** — removed from `remote.go`, `providers.go`, `common/providers.go`, `provider_config.go`, `resource_types.go`
