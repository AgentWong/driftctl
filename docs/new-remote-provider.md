# Remote provider architecture (deprecated)

> **Note:** As of v1.0.0, driftctl supports only the **AWS** provider. The multi-provider architecture (Azure, GCP, GitHub) has been removed. This document is retained for historical reference only.

## Current architecture

driftctl now uses a single `BulkEnumerator` backed by AWS Config to discover all AWS resources in one API call. The provider setup in `enumeration/remote/aws/init.go` creates:

1. An AWS Terraform provider (for state reading via gRPC)
2. A `ConfigRepository` (wrapping the AWS Config SDK with pagination and caching)
3. A `ConfigEnumerator` (implementing `BulkEnumerator`, mapping Config results to Terraform types)

The `ConfigEnumerator` is registered via `remoteLibrary.AddBulkEnumerator()` and replaces the former 103 individual `AddEnumerator()` calls.

## Adding support for new resource types

To add a new AWS resource type, see [Adding new resource type mappings](new-resource.md). No new enumerator or repository code is needed — only a mapping entry in `config_resource_mapping.go`.
