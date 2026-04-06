# Code Logic Diagrams

This document captures the logical flow of driftctl — both how the **legacy** upstream codebase (snyk/driftctl) worked, and how the **current** fork works today.

---

## Legacy Workflow (snyk/driftctl)

### Overview

The original codebase supported four cloud providers (AWS, Azure, GCP, GitHub) with 103+ individual per-resource-type enumerators for AWS alone. Drift detection was purely state-based: compare what Terraform state says should exist against what the cloud provider actually has.

### Top-Level Scan Flow

```mermaid
flowchart TD
    A([CLI: driftctl scan]) --> B[Parse flags & config]
    B --> C[Initialise providers<br>AWS · Azure · GCP · GitHub]
    C --> D[Read Terraform state<br>iacSupplier.Resources]
    D --> E[Scan cloud resources<br>Scanner.Resources]
    E --> F[Apply middleware chain<br>37+ AWS normalisation transforms]
    F --> G[Apply filters<br>driftignore · JMESPath]
    G --> H[Analyze<br>Analyzer.Analyze]
    H --> I[Output results<br>console · JSON · HTML]
    I --> J([Exit 0 = in sync<br>Exit 1 = drift detected])
```

### Enumeration: Per-type Enumerators

```mermaid
flowchart TD
    Scanner --> Pool[Goroutine pool<br>max 10 concurrent]
    Pool --> E1[S3BucketEnumerator<br>→ S3Repository]
    Pool --> E2[EC2InstanceEnumerator<br>→ EC2Repository]
    Pool --> E3[IAMRoleEnumerator<br>→ IAMRepository]
    Pool --> EN[... 100+ more enumerators]
    E1 & E2 & E3 & EN --> Collect[Collect<br>*Resource slice]
    Collect --> Return([Return all resources])
```

Each enumerator:
1. Calls one AWS API (e.g. `ListBuckets`, `DescribeInstances`)
2. Converts raw SDK types to `*resource.Resource`
3. Returns results into the shared pool

### State Reading

```mermaid
flowchart LR
    iacSupplier --> |scheme detection| Local[tfstate://<br>tfstate+file://]
    iacSupplier --> S3[tfstate+s3://]
    iacSupplier --> HTTP[tfstate+http://]
    iacSupplier --> TFC[tfstate+tfcloud://]
    Local & S3 & HTTP & TFC --> Parser[statefilev4 parser<br>hashicorp/terraform]
    Parser --> Resources([resource.Resource slice])
```

Lock file (`.terraform.lock.hcl`) is read to detect the provider version in use.

### Analysis (State vs Cloud)

```mermaid
flowchart TD
    A[State resources] --> Analyzer
    B[Cloud resources] --> Analyzer
    Analyzer --> |for each state resource| Check{Found in cloud?}
    Check --> |yes| Managed[managed]
    Check --> |no| Deleted[deleted]
    B --> |remaining after match| Unmanaged[unmanaged]
    Managed & Deleted & Unmanaged --> Analysis([Analysis result])
```

**Output categories:**
- `managed` — resource exists in both state and cloud (in sync)
- `deleted` — resource in state but missing from cloud
- `unmanaged` — resource in cloud but not tracked in state

---

## Current Workflow (this fork, v1.0.0)

### What Changed at a Glance

| Area | Legacy | Current |
|---|---|---|
| Providers | AWS · Azure · GCP · GitHub | **AWS only** |
| Enumeration model | 103 individual enumerators | 1 BulkEnumerator (AWS Config SQL only — no fallback) |
| Scan modes | Inventory only | **Inventory** (default) + **Plan** |
| Drift granularity | Resource existence only | Existence + **attribute-level diffs** (plan mode) |
| Resource categories | None | `cloudformation_managed` · `service_linked` · `unsupported` · `default_resources` |
| AWS SDK | v1 | **v2** (smithy-go errors, paginator APIs) |
| Terraform provider default | 3.19.0 | **6.38.0** |

### Top-Level Scan Flow (mode selection)

```mermaid
flowchart TD
    A([CLI: driftctl scan]) --> B[Parse flags & config]
    B --> ModeCheck{--mode flag}
    ModeCheck --> |inventory<br>default| Inv[Inventory mode]
    ModeCheck --> |plan| Plan[Plan mode]

    Inv --> D[Read Terraform state<br>iacSupplier.Resources]
    D --> E[Scan cloud — BulkEnumerator<br>AWS Config SQL only]
    E --> F[Apply middleware chain]
    F --> G[Apply driftignore / JMESPath filters]
    G --> H[Analyzer.Analyze<br>state vs cloud]
    H --> Cat[Categorize unmanaged resources]
    Cat --> ExCat[Filter by --exclude-category]
    ExCat --> Out[Output]

    Plan --> P1[Run terraform init + plan + show<br>terraform-exec]
    P1 --> P2[ParsePlan<br>extract resource changes + attribute diffs]
    P2 --> P3[Scan cloud — BulkEnumerator<br>no state reading]
    P3 --> P4[PlanAnalyzer.Analyze<br>plan vs cloud]
    P4 --> Out

    Out --> J([Exit 0 = in sync<br>Exit 1 = drift detected])
```

### Enumeration: AWS Config Only

```mermaid
flowchart TD
    Scanner --> Config[ConfigEnumerator<br>AWS Config SelectResourceConfig SQL]
    Config --> Batch[Batch up to 50 types<br>per SQL query]
    Batch --> API[AWS Config API<br>returns 132 resource types]
    API --> AllResources([All resources])
```

AWS Config covers 132 resource types via a single fast SQL query (enumeration time: ~2 s). There is no individual-enumerator fallback — if a resource type is not indexed by AWS Config, it is not enumerated. This keeps the required IAM permissions minimal: read access to AWS Config and S3 state buckets is sufficient.

### AWS Config Repository

```mermaid
sequenceDiagram
    participant CE as ConfigEnumerator
    participant CR as ConfigRepository
    participant AWS as AWS Config API

    CE->>CR: ListAllDiscoveredResources(configTypes[132])
    loop Chunks of 50 types
        CR->>AWS: SelectResourceConfig(SQL expression)
        AWS-->>CR: Page of ConfigurationItem records
        CR->>CR: paginate until complete
    end
    CR-->>CE: []DiscoveredResource
    CE->>CE: Map Config type → Terraform resource type
    CE-->>Scanner: []*resource.Resource
```

### Inventory Mode Analysis

```mermaid
flowchart TD
    StateRes[State resources] --> Analyzer
    CloudRes[Cloud resources] --> Analyzer
    Analyzer --> |for each state resource| Check{Found in cloud?}
    Check --> |yes| Managed[managed]
    Check --> |no| Deleted[deleted]
    CloudRes --> |remaining after match| Unmanaged[unmanaged]
    Unmanaged --> Categorizer

    Categorizer --> CF{CloudFormation API<br>physical ID lookup?}
    CF --> |yes| CFManaged[cloudformation_managed]
    CF --> |no| DR{default resource?<br>event bus · SSO role<br>KMS alias/aws/* etc.}
    DR --> |yes| DefManaged[default_resources]
    DR --> |no| SL{service-linked<br>role pattern?}
    SL --> |yes| SLManaged[service_linked]
    SL --> |no| US{type in Config<br>mapping?}
    US --> |no| Unsupported[unsupported]
    US --> |yes| StillUnmanaged[unmanaged]

    Managed & Deleted & CFManaged & DefManaged & SLManaged & Unsupported & StillUnmanaged --> ExcludeFilter[--exclude-category filter]
    ExcludeFilter --> Analysis([Analysis result])
```

### Plan Mode Analysis

```mermaid
flowchart TD
    TF[Terraform module dirs<br>--terraform-dir] --> MR[plan/multi_runner.go<br>RunParallel]

    subgraph Phase1["Phase 1 — sequential init (avoids SSO token file races)"]
        direction LR
        I1[terraform init<br>module 1] --> I2[terraform init<br>module 2] --> IN[...]
    end

    subgraph Phase2["Phase 2 — parallel plan"]
        direction LR
        P1[terraform plan + show<br>module 1]
        P2[terraform plan + show<br>module 2]
        PN[...]
    end

    MR --> Phase1
    Phase1 --> Phase2
    Phase2 --> Parser[plan/parser.go<br>ParsePlan per module]
    Parser --> DriftResult[DriftResult<br>resource changes + attribute diffs]

    DriftResult --> PA[PlanAnalyzer.Analyze]
    CloudInv[Cloud inventory<br>ConfigEnumerator] --> PA

    PA --> |plan action = update| Drifted[drifted<br>with AttributeChanges]
    PA --> |plan action = delete| PDeleted[deleted]
    PA --> |plan action = no-op| PManaged[managed]
    PA --> |in cloud, not in plan| PUnmanaged[unmanaged]

    Drifted & PDeleted & PManaged & PUnmanaged --> Output([Analysis result<br>with attribute-level diffs])
```

### Categorizer Chain

```mermaid
flowchart LR
    Resource --> C1[CloudFormationCategorizer]
    C1 --> |physical ID in stack| CFC([cloudformation_managed])
    C1 --> |no match| C2[DefaultResourceCategorizer]
    C2 --> |event bus · SSO role · KMS alias/aws/*| DRC([default_resources])
    C2 --> |no match| C3[ServiceLinkedCategorizer]
    C3 --> |path/name match| SLC([service_linked])
    C3 --> |no match| C4[UnsupportedCategorizer]
    C4 --> |not in Config mapping| UC([unsupported])
    C4 --> |in Config mapping| Default([unmanaged])
```

### Output Model

```mermaid
classDiagram
    class Analysis {
        +managed []*Resource
        +deleted []*Resource
        +unmanaged []*Resource
        +drifted []*DriftedResource
        +IsSync() bool
    }
    class DriftedResource {
        +Res *Resource
        +AttributeChanges []AttributeChange
    }
    class AttributeChange {
        +Path string
        +Before any
        +After any
    }
    class Summary {
        +TotalManaged int
        +TotalDeleted int
        +TotalUnmanaged int
        +TotalDrifted int
    }
    Analysis --> DriftedResource
    DriftedResource --> AttributeChange
    Analysis --> Summary
```

---

## Side-by-Side: Legacy vs Current

```mermaid
flowchart LR
    subgraph Legacy["Legacy (snyk/driftctl)"]
        direction TB
        L1[4 providers<br>AWS · Azure · GCP · GitHub]
        L2[103+ individual enumerators<br>one API call per resource type]
        L3[Read Terraform state<br>required]
        L4[37+ middlewares<br>normalisation]
        L5[Analyzer<br>state vs cloud]
        L6[3 categories<br>managed · deleted · unmanaged]
        L7[Resource existence only<br>no attribute diffs]
        L1 --> L2 --> L3 --> L4 --> L5 --> L6 --> L7
    end

    subgraph Current["Current fork (v1.0.0)"]
        direction TB
        C1[AWS only<br>KISS principle]
        C2[1 BulkEnumerator<br>AWS Config SQL only<br>132 types · 2 s · no fallback]
        C3[State OR Terraform plan<br>plan mode needs no state]
        C4[Same 37+ middlewares<br>inventory mode only]
        C5[Analyzer or PlanAnalyzer<br>choice per mode]
        C6[7 categories<br>managed · deleted · drifted · unmanaged<br>cloudformation_managed · default_resources · service_linked · unsupported]
        C7[Attribute-level diffs<br>plan mode]
        C1 --> C2 --> C3 --> C4 --> C5 --> C6 --> C7
    end
```

---

## Key File Reference

| Concern | Legacy path | Current path |
|---|---|---|
| CLI entry | `pkg/cmd/scan.go` | `pkg/cmd/scan.go` |
| Orchestration | `pkg/driftctl.go` | `pkg/driftctl.go` |
| Scanner | `enumeration/remote/scanner.go` | `enumeration/remote/scanner.go` |
| AWS init | `enumeration/remote/aws/init.go` (103 AddEnumerator) | `enumeration/remote/aws/init.go` (1 AddBulkEnumerator) |
| AWS enumeration | 103 `*_enumerator.go` files | `enumeration/remote/aws/config_enumerator.go` |
| AWS repository | 20+ `*_repository.go` files | `enumeration/remote/aws/repository/config_repository.go` |
| CloudFormation repository | — | `enumeration/remote/aws/repository/cloudformation_repository.go` |
| Config mapping | — | `enumeration/remote/aws/config_resource_mapping.go` |
| State reading | `pkg/iac/supplier/` | `pkg/iac/supplier/` (unchanged) |
| Middlewares | `pkg/middlewares/` | `pkg/middlewares/` (unchanged) |
| Analyzer | `pkg/analyser/analyzer.go` | `pkg/analyser/analyzer.go` |
| Plan analyzer | — | `pkg/analyser/plan_analyzer.go` |
| Plan runner | — | `pkg/terraform/plan/runner.go` |
| Plan parser | — | `pkg/terraform/plan/parser.go` |
| Categorizer | — | `pkg/categorizer/` |
| Analysis model | `pkg/analyser/analysis.go` | `pkg/analyser/analysis.go` (extended) |
