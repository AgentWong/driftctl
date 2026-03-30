# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-03-30

### Added
- **golangci-lint — enabled 11 new linters** in `.golangci.yml`: `bodyclose`, `dupword`, `durationcheck`, `errorlint`, `gosec`, `misspell`, `nilerr`, `unconvert`, `usestdlibvars`, `wastedassign`, `whitespace`. Configured `gosec` to exclude G101 false positives and added `linters.exclusions.rules` for test-only G304/G703 suppressions.

### Fixed
- **errorlint** — replaced direct type assertions on errors with `errors.As`/`errors.Is` in `enumeration/remote/resource_enumeration_error_handler.go`, `enumeration/terraform/provider_installer.go`, `main.go`, `pkg/cmd/scan.go`, `pkg/iac/terraform/state/terraform_state_reader.go`, `pkg/middlewares/aws_api_gateway_api_expander.go`, and `test/acceptance/testing.go`
- **nilerr** — fixed 4 instances in `pkg/middlewares/aws_api_gateway_api_expander.go` where errors were silently swallowed (`return nil` when `err != nil`), now properly returns the error
- **bodyclose** — closed HTTP response bodies in `pkg/telemetry/telemetry.go` and `pkg/version/version.go`; added nolint for `http_reader.go` where the struct manages body lifecycle
- **misspell** — corrected "occured" → "occurred" in `enumeration/remote/alerts/alerts.go` and corresponding test data
- **usestdlibvars** — replaced string/int literals with `http.MethodGet`, `http.MethodPost`, `http.StatusOK`, `http.StatusNotFound` across 4 files
- **whitespace** — removed unnecessary leading/trailing newlines across ~25 files
- **gosec** — tightened file permissions (`os.ModePerm` → `0750`/`0600`) in `test/goldenfile/goldenfile.go`, `test/schemas/shemas.go`, `test/files.go`, and `enumeration/terraform/provider_installer_test.go`; added targeted nolint annotations for legitimate false positives (trusted embedded assets, plugin paths, test fixtures)

### Changed
- **AWS SDK v1 → v2 migration (test infrastructure)** — migrated all remaining AWS SDK v1 (`aws-sdk-go`) usage to v2 (`aws-sdk-go-v2`) in test files. Replaced `session.Session` with `aws.Config` in `test/acceptance/awsutils/aws.go`, updated 6 acceptance tests in `pkg/resource/aws/` and 1 in `pkg/iac/terraform/state/` to use v2 client constructors (`NewFromConfig`), v2 types (e.g., `ec2types.Filter`, `ecrtypes.ImageTagMutabilityImmutable`), and `context.Context` parameters. Deleted 37 dead files in `test/aws/` (22 auto-generated mocks and 15 interface definitions) that were unused outside the directory. Zero direct v1 imports remain in the codebase.

### Fixed
- **HTML output template** — updated `pkg/cmd/scan/output/assets/index.tmpl` to call `ResourceID` (7 occurrences) instead of the old `ResourceId` method name, resolving a template execution error when writing the HTML report

### Changed
- **Linting — `revive` compliance (var-naming, unused-parameter, error-strings, exported)**
  - `var-naming`: renamed `Resource.Id` → `Resource.ID`, `SerializableResource.Id` → `SerializableResource.ID`, and `ResourceId()` → `ResourceID()` across all callers (~200 references in 50+ files)
  - `unused-parameter`: replaced all unused function parameters with `_` in test callbacks and cobra command handlers across `pkg/cmd/`, `pkg/driftctl_test.go`, `pkg/middlewares/`, `pkg/resource/aws/`, and `pkg/telemetry/`
  - `error-strings`: lowercased error strings passed to `errors.New()` in `enumeration/remote/resource_enumeration_error_handler_test.go`; added `//nolint:revive` for one string that mirrors an external Terraform error message verbatim
  - `exported`: added doc comment for `Chain` type in `pkg/middlewares/chain_middleware.go`; rewrote `AwsNatGatewayEipAssoc.Execute` comment to begin with the method name per godoc convention

## [1.0.0] - 2026-03-29

### Changed
- **CloudFormation categorization — replaced heuristic regex/tag matching with authoritative CloudFormation API** in `pkg/categorizer/cloudformation.go`. Previously used three unreliable strategies: `aws:cloudformation:stack-name` tag inspection (fails for untaggable resource types), regex matching on physical ID naming conventions (`<stack>-<LogicalId>-<12-13 char suffix>`), and CDK/AwsSolutions path patterns. Now calls `ListStacks` + `ListStackResources` via new `CloudFormationRepository` to build a definitive set of physical resource IDs that CloudFormation manages, then does a simple set-membership lookup. This eliminates false positives (e.g., resources with names that happened to match the regex) and false negatives (e.g., untaggable resources or stacks with non-standard naming).
- **New `CloudFormationRepository`** in `enumeration/remote/aws/repository/cloudformation_repository.go` — queries all active CloudFormation stacks (CREATE_COMPLETE, UPDATE_COMPLETE, etc.) and enumerates their resources via `ListStackResources`, returning a cached set of physical resource IDs
- **`aws.Init` and `remote.Activate` now return `aws.Config`** — allows `scan.go` to create additional AWS service clients (e.g., CloudFormation) after provider initialization without threading the config through unrelated layers

### Added
- **AWS Config query now returns ARN, tags, and resource name** in `config_repository.go` — the `SelectResourceConfig` SQL expression now selects `arn` and `tags` in addition to the existing fields, enabling downstream categorization and richer display
- **Config enumerator populates resource attributes** in `config_enumerator.go` — tags, ARN, and Config resource name are seeded into `CreateAbstractResource` so categorizers and output formatters can use them
- **Resource `Name` field in JSON output** — `SerializableResource` now includes a `name` field derived from Config's `resourceName` or the `Name` tag; `Resource.DisplayName()` method added for template use
- **Managed Resources tab in HTML report** — new tab displays Terraform-managed resources with ID, Type, Name, and IaC Source columns
- **CloudFormation Managed tab in HTML report** — unmanaged resources categorized as `cloudformation_managed` are now shown in a dedicated tab, separate from other unmanaged resources
- **Name column in all HTML report tables** — all resource tables (Managed, Unmanaged, CloudFormation Managed, Missing, Drifted) now include a Name column for human-readable identification
- **Default Resources category** — new categorizer detects AWS auto-created resources (default event buses, managed event rules like `AutoScalingManagedRule` and `IS-Tagging-default-*`, SSO reserved roles `AWSReservedSSO_*`, and default KMS aliases `alias/aws/*`); these are shown in a dedicated "Default Resources" tab and excluded from unmanaged totals
- **CloudFormation name-based detection** — the CloudFormation categorizer now also matches resources by the CloudFormation physical-ID naming convention (`<stack>-<LogicalId>-<12-char suffix>`), catching CloudFormation-managed resources that lack the `aws:cloudformation:stack-name` tag (e.g., `aws-instance-scheduler-*` resources)
- **Selective JSON report parsing instructions** in `.github/agents/build-test-fix.md` — added Step 3a with `jq`-based parsing patterns to avoid overwhelming LLM context with large reports

### Changed
- **AWS Terraform provider default** — updated from `5.82.2` to `6.38.0` in `enumeration/remote/aws/provider.go` and `pkg/resource/schemas/repository.go`
- **HTML report search** — search box now matches against both Resource ID and Name fields
- **HTML report tab initialization** — JavaScript now selects the first available tab on page load, fixing edge cases where no tab was initially active
- **HTML report table layout** — switched from flex-based row layout to proper `<table>` display with column borders, header styling, and fixed-width columns for clearer visual separation between Name, Resource Type, and Resource ID fields

### Changed
- **Config enumeration — switched to SelectResourceConfig (Advanced Query) API** in `enumeration/remote/aws/repository/config_repository.go`. Replaces the previous `ListDiscoveredResources` + `GetDiscoveredResourceCounts` approach, which lagged on newly-started Config recorders. The new API queries the Config resource index directly via SQL expressions, returning results immediately and reducing scan time (8s → 2s for enumeration). Types are batched into chunks of 50 to stay within the 4 KB SQL expression limit.
- **Removed `GetSupportedResourceTypes()`** from `ConfigRepository` interface — no longer needed since the enumerator passes its known mapping keys directly.
- **Config enumerator passes resource types explicitly** in `enumeration/remote/aws/config_enumerator.go` — extracts Config type keys from the mapping and passes them to `ListAllDiscoveredResources` rather than relying on the repository to discover types.

### Fixed
- **Test suite — all 26 packages now pass** after fixing test expectations broken by prior refactoring:
  - `pkg/cmd/scan_test.go`, `pkg/cmd/driftctl_test.go`, `pkg/iac/supplier/supplier_test.go` — removed `gs://` (GCS) from expected scheme/backend error messages
  - `pkg/cmd/completion_test.go` — updated expected zsh and powershell completion output for newer cobra v1.10.2
  - `pkg/iac/terraform/state/versions_test.go` — updated error message casing (`malformed` vs `Malformed`) for hashicorp/go-version v1.8.0
  - `pkg/filter/driftignore_test.go` — replaced `azurerm_route_table`/`azurerm_route` with `aws_route_table`/`aws_route` (AWS-only scope)
  - `enumeration/remote/resource_enumeration_error_handler_test.go` — rewrote from v1 `awserr` to v2 smithy-go error types
  - `pkg/iac/terraform/state/backend/s3_reader_test.go` — introduced `S3GetObjectAPI` interface and local mock, replacing v1 S3 mock
  - `pkg/iac/terraform/state/enumerator/s3_test.go` — introduced `mockListObjectsV2Client` using v2 `ListObjectsV2APIClient` interface, replacing v1 pagination callback mocks
- **Categorizer — `AWSServiceRoleFor*` IAM roles classified as default resources** — moved `AWSServiceRoleFor` prefix detection from `ServiceLinkedCategorizer` to `DefaultResourceCategorizer` and added a resource-ID fallback so Config-enumerated service-linked roles are no longer misclassified as unmanaged
- **Categorizer — `resourceName()` now checks `config_name` attribute** — Config-enumerated resources store their AWS Config name as `config_name`, not `name`; the helper now falls back to `config_name`, fixing name-based pattern matching for all categorizers
- **Categorizer — CloudFormation regex no longer false-positives on UUIDs and ARNs** — KMS key UUIDs and ACM certificate ARNs contain 12-hex-char segments that matched the old `{12}` suffix pattern; added UUID and ARN exclusion to `matchesCfnNamePattern`
- **Categorizer — CloudFormation suffix expanded to 12–13 characters** — recent CloudFormation stacks can generate 13-char random suffixes; `cfnPhysicalIDPattern` regex updated from `{12}` to `{12,13}`
- **Categorizer — CDK/AWS Solutions naming pattern detection** — added `cdkNamePattern` regex to match `<lowercase-prefix>-<CamelCaseLogicalId>` resource names emitted by CDK and AWS Solutions stacks (e.g., `default-Ec2ResizeRequestHandler-Role`)
- **Categorizer — `AwsSolutions` path detected as CloudFormation** — KMS aliases and other resources containing `/AwsSolutions/` in their name or ID are now classified as `cloudformation_managed`
- **Test golden files — added `total_default_resources` to JSON/HTML expectations** — updated 5 JSON golden files and 4 HTML golden files to include the new `total_default_resources` summary field added by the categorizer changes

### Changed
- **S3 backend testability** — `s3_reader.go` `S3Client` field changed from `*s3.Client` to `S3GetObjectAPI` interface; `s3.go` (enumerator) `client` field changed from `*s3.Client` to `s3.ListObjectsV2APIClient` interface
- **Provider download — native arm64 support** — removed outdated `darwin/arm64 → amd64` architecture override in `provider_config.go`; modern provider versions ship native arm64 binaries
- **AWS SDK v1 import cleanup** — migrated ~35 test files from v1 `aws-sdk-go/aws` pointer helpers and `awsutil.Prettify` to v2 equivalents (`aws-sdk-go-v2/aws` and `fmt.Sprintf`)
- **Linting — golangci-lint v2 enabled with `staticcheck`, `gocritic`, and `revive`** — `.golangci.yml` updated to version 2 format with three additional linters:
  - `staticcheck`: `WriteString(Sprintf(...))` → `fmt.Fprintf`, removed deprecated `cobra.ExactValidArgs`, removed unnecessary `.Analysis.` embedded field selectors, lowercased error strings, lifted loop conditions
  - `gocritic`: `else { if }` → `else if`, if-else chains → `switch`, `+= ` assignment operators, `<= 0` → `== 0`, removed redundant unslice `[:]`, dereferenced pointer receivers simplified
  - `revive`: added doc comments to all exported constants, types, and functions across `enumeration/resource/aws/`, `pkg/`, and `test/` packages; renamed `EXIT_IN_SYNC/EXIT_NOT_IN_SYNC/EXIT_ERROR` → `ExitInSync/ExitNotInSync/ExitError`; renamed `NormalizeJsonString` → `NormalizeJSONString`; added package comments
  - `errcheck`: all unchecked `Close()`, `os.Remove()`, `os.Setenv()`, and `fmt.Fprintf()` return values explicitly discarded with `_ =` or wrapped in `defer func() { _ = ... }()`

### Refactored — `revive` lint compliance (bulk)
- **Stuttering type renames** — renamed types whose package-qualified names stuttered:
  - `build.BuildInterface` → `build.Interface`
  - `parallel.ParallelRunner` → `parallel.Runner` (and `NewParallelRunner` → `NewRunner`)
  - `terraform.TerraformProvider` → `terraform.Provider` (in `enumeration/terraform/`)
  - `filter.FilterEngine` → `filter.Engine`
  - `output.OutputConfig` → `output.Config`
  - `remote/terraform.TerraformProvider` → `remote/terraform.Provider` and `TerraformProviderConfig` → `Config`
- **`var-naming` const renames** — renamed API Gateway constants to use idiomatic Go initialisms:
  - `AwsApiGateway*` → `AwsAPIGateway*` (Account, BasePathMapping, Deployment, RequestValidator, RestApi→RestAPI, VpcLink, V2RouteResponse, and others)
  - `AwsApiGatewayV2*` func/init names updated to match (`initAwsApiGatewayV2*` → `initAwsAPIGatewayV2*`)
- **`var-naming` field/param renames** — `Id` → `ID`, `tableId` → `tableID`, `PrefixListId` → `PrefixListID`, `networkAclId` → `networkACLID`, `CidrBlock` → `cidrBlock` (unexported params), `GetDownloadUrl` → `GetDownloadURL`, `NormalizeJsonString` → `NormalizeJSONString`, `CreateSecurityGroupRuleIdHash` → `CreateSecurityGroupRuleIDHash`
- **`exported` doc comments** — added doc comments to all exported types, functions, methods, and constants across:
  - `build/`, `enumeration/alerter/`, `enumeration/diagnostic/`, `enumeration/parallel/`, `enumeration/terraform/`, `enumeration/remote/`
  - `pkg/analyser/`, `pkg/categorizer/`, `pkg/cmd/`, `pkg/filter/`, `pkg/helpers/`, `pkg/output/`, `pkg/resource/`, `pkg/version/`
  - `sentry/`, `test/remote/`, `test/tfe/`
  - ~70 `pkg/resource/aws/` resource type constant files
  - ~50 `enumeration/resource/aws/` resource type constant files
- **`package-comments`** — added package-level doc comments to `filter`, `helpers`, `diagnostic`, `sentry`, `terraform` (enumeration), `remote` (test), `lock`, `output`, `version`, `hcl`, `backend`, and `aws` (pkg/resource)
- **`unused-parameter`** — replaced unused parameters with `_` in test callbacks (`provider_downloader_test.go`, `driftctl_test.go`, `provider_installer.go`, `aws/init.go`)
- **`receiver-naming`** — standardized receiver names for `AWSTerraformProvider` (all use `a`)
- **`unexported-return`** — changed `NewProgress` return type from `*progress` to `Progress` interface; changed `NewTFCloudConfigReader` to return exported `TFCloudConfigReader` type
- **`empty-block`** — added `runtime.Gosched()` to busy-wait loop in `progress_test.go`
- **`indent-error-flow`** — removed unnecessary `else` after early-return `if` blocks in middleware files

## [1.0.0] - 2026-03-28

### Changed
- **Go version** — bumped from 1.23 to 1.24+ (go.mod, .go-version, Dockerfile)
- **AWS Terraform provider default** — updated from `3.19.0` to `5.82.2` in `enumeration/remote/aws/provider.go` and `pkg/resource/schemas/repository.go`
- **AWS SDK v1 → v2 migration (core production code)** — migrated the following to `aws-sdk-go-v2`:
  - `enumeration/remote/aws/provider.go` — replaced `session.Session` with `aws.Config`, `sts.New()` with `sts.NewFromConfig()`
  - `enumeration/remote/aws/repository/config_repository.go` — replaced `configservice` v1 client with v2 paginator-based API
  - `enumeration/remote/aws/client/s3_client_factory.go` — replaced `client.ConfigProvider` with `aws.Config`, v2 service constructors
  - `pkg/iac/terraform/state/backend/s3_reader.go` — replaced v1 session/S3 client with v2 `config.LoadDefaultConfig`/`s3.NewFromConfig`
  - `pkg/iac/terraform/state/enumerator/s3.go` — replaced v1 session/pagination with v2 config/paginator
  - `enumeration/remote/resource_enumeration_error_handler.go` — replaced `awserr.RequestFailure` with smithy-go error types
  - `pkg/middlewares/aws_s3_bucket_public_access_block_reconcilier.go` — replaced `aws.BoolValue` with `aws.ToBool`
  - **Note:** ~86 test/mock files still reference AWS SDK v1 and need migration (left for iterative testing)
- **`hashicorp/terraform` v0.14.0 partial replacement** — replaced `helper/hashcode` usage across 6 files with local `pkg/helpers/hashcode.go` implementation. Remaining `plugin`, `providers`, `states`, and `addrs` imports are deeply architectural (provider gRPC protocol, state file parsing) and require a larger refactoring effort
- **`hashicorp/hc-install` adoption** — replaced deprecated `terraform-exec/tfinstall` package in `test/acceptance/testing.go` with `hc-install` v0.9.3 (`releases.ExactVersion`, `fs.AnyVersion`)
- **`go-hclog` v1.6.3 interface compliance** — added `GetLevel()` method to `TerraformPluginLogger` in `logger/plugin_logger.go`

### Updated Dependencies
| Dependency | From | To |
|---|---|---|
| `spf13/cobra` | v1.0.0 | v1.10.2 |
| `spf13/viper` | v1.7.1 | v1.21.0 |
| `spf13/pflag` | v1.0.5 | v1.0.10 |
| `getsentry/sentry-go` | v0.10.0 | v0.44.1 |
| `hashicorp/terraform-json` | v0.12.0 | v0.27.2 |
| `hashicorp/terraform-exec` | v0.14.0 | v0.25.0 |
| `hashicorp/hcl/v2` | v2.7.2 | v2.24.0 |
| `hashicorp/go-hclog` | v0.9.2 | v1.6.3 |
| `hashicorp/go-version` | v1.6.0 | v1.8.0 |
| `go-git/go-git/v5` | v5.4.2 | v5.17.0 |
| `fatih/color` | v1.9.0 | v1.19.0 |
| `sirupsen/logrus` | v1.9.3 | v1.9.4 |
| `stretchr/testify` | v1.8.3 | v1.11.1 |
| `zclconf/go-cty` | v1.8.4 | v1.18.0 |
| `r3labs/diff/v2` | v2.6.0 | v2.15.1 |
| `bmatcuk/doublestar/v4` | v4.0.1 | v4.10.0 |
| `eapache/go-resiliency` | v1.3.0 | v1.7.0 |
| `joho/godotenv` | v1.3.0 | v1.5.1 |

### Removed
- **GCS backend support** — removed `gs_reader.go`, `gs.go` (enumerator), GCS backend case from `backend.go`, `state_enumerator.go`, and `hcl/backend.go`
- **Azure backend HCL test fixtures** — removed `azurerm_backend_block.tf`, `azurerm_backend_workspace/`, `gcs_backend_block.tf`
- **Azure/GCP/GitHub resource type definitions** — removed 75+ entries from `resource_types.go`
- **Azure/GCP/GitHub test data** — removed ~63 test directories (`pkg/iac/terraform/state/test/azurerm_*`, `google_*`, `github_*`, `enumeration/remote/test/azurerm_*`, etc.)
- **Azure/GCP/GitHub test schemas** — removed `pkg/test/azurerm/`, `pkg/test/github/`, `pkg/test/google/`
- **Azure helper code** — removed `pkg/helpers/azure/`
- **GCP test helpers** — removed `test/google/`
- **Direct GCP/Azure SDK dependencies** — removed `cloud.google.com/go/asset`, `cloud.google.com/go/storage`, 4 `Azure/azure-sdk-for-go` modules, `google.golang.org/api`, `google.golang.org/grpc` from direct deps
- **Deprecated indirect deps** — `golang/mock`, `go-checkpoint`, various unused transitive deps cleaned via `go mod tidy`

### Known Issues / Remaining Work
- **Test schema fixtures** — `pkg/test/aws/3.19.0/schema.json` still contains v3.19.0 provider schemas; tests reference this version. A new `5.82.2/schema.json` needs to be generated by running the provider and capturing its schema output
- **AWS SDK v1 in tests** — ~86 test and mock files still import `aws-sdk-go` v1. These need migration to v2 and mock regeneration
- **`hashicorp/terraform` v0.14.0** — still required for `plugin` (gRPC provider communication), `providers` (schema types), `states`/`statefile` (state parsing), and `addrs` (resource addressing). Full replacement requires rewriting the provider plugin layer using `terraform-plugin-go`
- **`hashicorp/go-tfe`** — still at v0.20.0 (current is v1.x); used in `tfcloud_reader.go` for Terraform Cloud state reading

## [1.0.0] - 2026-03-27

### Added
- **AWS Config-based inventory system** — replaced 103 individual AWS resource enumerators with a single AWS Config API integration (`config_repository.go`, `config_enumerator.go`, `config_resource_mapping.go`) covering 132 resource type mappings
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
- **Azure provider** (~241 files) — all `azurerm` enumerators, repositories, resource types, middlewares, state backend readers, and test schemas
- **Google Cloud provider** (~220 files) — all `google` enumerators, repositories, resource types, middlewares, and test schemas
- **GitHub provider** (~36 files) — all `github` enumerators, repositories, resource types, and test schemas
- **103 individual AWS enumerators** — replaced by the Config-based BulkEnumerator; all `*_enumerator.go` files removed (except `config_enumerator.go`)
- **65 individual AWS repository, mock, and test files** — replaced by `config_repository.go`; old `ec2_repository.go`, `s3_repository.go`, `iam_repository.go`, etc. plus associated mocks and tests removed
- **20 AWS scanner test files** — all `aws_*_scanner_test.go` files for deleted enumerators
- **Non-AWS provider constants and references** — removed from `remote.go`, `providers.go`, `common/providers.go`, `provider_config.go`, `resource_types.go`, `backend.go`, `hcl/backend.go`, `state_enumerator.go`, `schemas/repository.go`, `alerts.go`, `resource_enumeration_error_handler.go`
- **Non-AWS SDK dependencies** — pruned 6 Azure SDK `resourcemanager` modules, `githubv4` GraphQL client, and transitive dependencies via `go mod tidy`
