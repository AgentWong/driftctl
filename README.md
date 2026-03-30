<p align="center">
  <img width="200" src="https://docs.driftctl.com/img/driftctl_dark.svg" alt="driftctl">
</p>

<p align="center">
  <img src="https://circleci.com/gh/snyk/driftctl.svg?style=shield"/>
  <img src="https://goreportcard.com/badge/github.com/snyk/driftctl"/>
  <img src="https://img.shields.io/github/license/snyk/driftctl">
  <img src="https://img.shields.io/github/v/release/snyk/driftctl">
  <img src="https://img.shields.io/github/go-mod/go-version/snyk/driftctl">
  <img src="https://img.shields.io/github/downloads/snyk/driftctl/total.svg"/>
  <img src="https://img.shields.io/docker/pulls/snyk/driftctl"/>
  <img src="https://img.shields.io/docker/image-size/snyk/driftctl"/>
  <a href="https://discord.gg/NMCBxtD7Nd">
    <img src="https://img.shields.io/discord/783720783469871124?color=%237289da&label=discord&logo=discord"/>
  </a>
</p>

<p align="center">
  Detect infrastructure drift and unmanaged AWS resources using AWS Config and Terraform.<br>
  <strong>IaC:</strong> Terraform. <strong>Cloud provider:</strong> AWS.<br>
</p>

<details>
  <summary>Packaging status</summary>
  <a href="https://repology.org/project/driftctl/versions">
    <img src="https://repology.org/badge/vertical-allrepos/driftctl.svg" alt="Packaging status">
  </a>
</details>

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

## Quick start

```shell
# Inventory mode: find unmanaged AWS resources
driftctl scan

# Plan mode: detect configuration drift + unmanaged resources
driftctl scan --mode plan --terraform-dir /path/to/terraform

# Exclude false positives by category
driftctl scan --exclude-category cloudformation_managed,default_resources,service_linked
```

## Links

**[Documentation](https://docs.driftctl.com)**

**[Installation](https://docs.driftctl.com/installation)**

**[Discord](https://discord.gg/7zHQ8r2PgP)**

## Contribute

To learn more about compiling driftctl and contributing, please refer to the [contribution guidelines](.github/CONTRIBUTING.md) and the [contributing guide](docs/README.md) for technical details.

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification and is brought to you by these [awesome contributors](CONTRIBUTORS.md).

Build with ❤️️ from 🇫🇷 🇬🇧 🇯🇵 🇬🇷 🇸🇪 🇺🇸 🇷🇪 🇨🇦 🇮🇱 🇩🇪

## Security notice

All Terraform state and Terraform files in this repository are for unit test
purposes only. No running code attempts to access these resources (except to
create and destroy them, in the case of acceptance tests). They are just opaque
strings.
