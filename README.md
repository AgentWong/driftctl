## Why driftctl?

Infrastructure drift is a blind spot and a source of potential security issues.
Drift can have multiple causes: from team members creating or updating infrastructure through the web console without backporting changes to Terraform, to unexpected actions from authenticated apps and services.

You can't efficiently improve what you don't track. We track coverage for unit tests, why not infrastructure as code coverage?

Spot discrepancies as they happen: driftctl is a free and open-source CLI that warns of infrastructure drifts and fills in the missing piece in your DevSecOps toolbox.

## How it works

driftctl offers two scan modes:

- **Inventory mode** (default) — Uses the [AWS Config](https://docs.aws.amazon.com/config/latest/developerguide/WhatIsConfig.html) API to discover all resources in your AWS account in a single API call, then compares them against your Terraform state to find unmanaged resources.
- **Plan mode** — Runs `terraform plan` against your Terraform root module to detect attribute-level configuration drift (e.g. a security group rule changed outside Terraform), then combines the results with AWS Config inventory to also identify unmanaged resources.

Resources are automatically categorized to reduce false positives:
- **CloudFormation-managed** — resources managed by CloudFormation stacks (detected via CloudFormation API, tags, and naming patterns)
- **Default resources** — AWS auto-created resources (default event buses, managed event rules, SSO reserved roles, default KMS aliases)
- **Service-linked** — AWS service-linked roles and resources
- **Unsupported** — resource types not covered by AWS Config

## Features

- **Inventory scan** — discover all AWS resources via AWS Config and compare against Terraform state
- **Plan-based drift detection** — detect attribute-level configuration drift using `terraform plan`
- **Resource categorization** — automatically classify resources to filter false positives
- **Multiple output formats** — console, JSON, and HTML reports with drift details
- Allow users to **ignore** resources via `.driftignore`
- **132 AWS resource types** supported via AWS Config mapping

## Directory Structure

```
driftctl/
├── bin/                        # Compiled binary
├── build/                      # Build version metadata
├── docs/                       # Developer guides (architecture, testing, adding resources)
├── logger/                     # Logging setup and formatters
├── mocks/                      # Auto-generated mock interfaces
├── test/                       # Test utilities, fixtures, and acceptance tests
│   ├── acceptance/             #   End-to-end acceptance tests
│   ├── goldenfile/             #   Golden file test helpers
│   ├── remote/                 #   Remote resource test fixtures
│   └── terraform/              #   Terraform state test fixtures
├── test-output/                # Test result reports (JSON, HTML)
│
├── enumeration/                # Resource discovery from AWS and Terraform state
│   ├── alerter/                #   Alert collection and reporting
│   ├── diagnostic/             #   Diagnostic information gathering
│   ├── parallel/               #   Parallel enumeration runner
│   ├── resource/               #   Resource abstraction and AWS type definitions
│   │   └── aws/                #     100+ AWS resource type definitions
│   ├── remote/                 #   Cloud provider enumeration
│   │   ├── aws/                #     AWS Config API enumerator and resource mapping
│   │   │   ├── client/         #       AWS SDK client setup
│   │   │   └── repository/     #       CloudFormation and Config resource repositories
│   │   ├── cache/              #     Enumeration result caching
│   │   └── common/             #     Shared enumeration interfaces
│   └── terraform/              #   Terraform provider management, state reading, schemas
│
└── pkg/                        # Core application logic
    ├── cmd/                    #   CLI commands (root, scan, completion, version)
    │   └── scan/               #     Scan command and output formatters (console, JSON, HTML)
    │       └── output/         #       Console, JSON, and HTML report renderers
    ├── analyser/               #   Drift analysis and Terraform plan comparison
    ├── categorizer/            #   Resource categorization (CloudFormation, defaults, service-linked)
    ├── filter/                 #   Resource filtering and .driftignore parsing
    ├── middlewares/             #   80+ AWS resource normalizers and reconcilers
    ├── resource/               #   Resource type constants, factories, and schemas
    │   └── aws/                #     AWS resource type definitions
    ├── iac/                    #   Infrastructure-as-Code state supplier and Terraform handling
    │   └── terraform/          #     Terraform-specific IaC integration
    ├── terraform/              #   Terraform plan execution and HCL parsing
    ├── config/                 #   Configuration file parsing
    ├── memstore/               #   In-memory resource store (buckets)
    ├── helpers/                #   Utility functions (hashing, JSON normalization)
    ├── http/                   #   HTTP client configuration
    ├── output/                 #   Progress reporting and output printer
    └── version/                #   Version constant
```

**Key flow:** `pkg/cmd/scan/` orchestrates the scan → `enumeration/` discovers resources from AWS Config and Terraform state → `pkg/middlewares/` normalizes resources → `pkg/analyser/` compares them → `pkg/categorizer/` classifies drift → `pkg/cmd/scan/output/` renders the report.

## Quick start

```shell
# Inventory mode: find unmanaged AWS resources
driftctl scan

# Plan mode: detect configuration drift + unmanaged resources
driftctl scan --mode plan --terraform-dir /path/to/terraform

# Exclude false positives by category
driftctl scan --exclude-category cloudformation_managed,default_resources,service_linked
```
