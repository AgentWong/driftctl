# Adding new resource type mappings

With the v1.0.0 refactoring, resource discovery is handled by a single AWS Config enumerator rather than individual per-resource enumerators. Adding support for a new AWS resource type is now primarily a mapping exercise.

## Prerequisites

The resource type must be [supported by AWS Config](https://docs.aws.amazon.com/config/latest/developerguide/resource-config-reference.html). If AWS Config doesn't track the resource type, it cannot be discovered by driftctl's inventory system.

## Step 1: Add the Config-to-Terraform mapping

Edit `enumeration/remote/aws/config_resource_mapping.go` and add an entry to the `configToTerraformType` map:

```go
var configToTerraformType = map[string]string{
    // ... existing mappings ...
    "AWS::NewService::ResourceType": "aws_new_resource",
}
```

The key is the [AWS Config resource type](https://docs.aws.amazon.com/config/latest/developerguide/resource-config-reference.html) and the value is the corresponding [Terraform resource type](https://registry.terraform.io/providers/hashicorp/aws/latest/docs).

## Step 2: Define the resource type constant

Add a file `enumeration/resource/aws/aws_new_resource.go` (if it doesn't already exist) with the type constant and optional metadata:

```go
const AwsNewResourceResourceType = "aws_new_resource"

func initAwsNewResourceMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
    // Optional: define normalize functions to exclude fields from comparison
    resourceSchemaRepository.SetNormalizeFunc(AwsNewResourceResourceType, func(res *resource.Resource) {
        val := res.Attrs
        val.SafeDelete([]string{"field_to_ignore"})
    })
}
```

Register the metadata init function in `enumeration/resource/aws/metadatas.go`:

```go
func InitResourcesMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
    // ... existing calls ...
    initAwsNewResourceMetaData(resourceSchemaRepository)
}
```

## Step 3: Register the type in supported types

Add the type to `pkg/resource/resource_types.go`:

```go
var supportedTypes = map[string]struct{}{
    // ... existing types ...
    "aws_new_resource": {},
}
```

## Step 4: Test

Run the scan to verify the new resource type is discovered:

```shell
driftctl scan --filter "Type=='aws_new_resource'"
```

## Notes

- No enumerator or repository code is needed — the `ConfigEnumerator` handles all Config-supported types automatically via the mapping table.
- If the resource requires special middleware for reconciliation (e.g. ID format differences between Config and Terraform), add a middleware in `pkg/middlewares/`.
- The mapping table currently has 132 entries. Run `UnsupportedByConfig()` to check which Terraform types lack Config coverage.
