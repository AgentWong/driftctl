# Middlewares

The main goal of middlewares is to reconciliate IaC and remote resources. For this we filter default resources, mutate, or remove fields in resources or even create and delete new resources.

Middlewares are only used in **inventory mode** (`--mode=inventory`). In **plan mode** (`--mode=plan`), the terraform plan output is already normalized and does not need middleware reconciliation.

```go
type AwsDefaultRoute struct{}

func NewAwsDefaultRoute() AwsDefaultRoute {
	return AwsDefaultRoute{}
}

func (m AwsDefaultRoute) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newRemoteResources := make([]*resource.Resource, 0)

	// ...

	*remoteResources = newRemoteResources

  return nil
}
```

In the above example, we define a middleware called `AwsDefaultRoute` that will modify remote resources. Middleware can access two arrays of type `*[]*resource.Resource`: IaC resources first and then remote resources. The goal is to rework these slices to remove false positive drifts. Notice both `remoteResources` and `resourcesFromState` variables are pointers, which mean middlewares can perform mutations on resources before the comparison is made.

## Different kind of middlewares

1) Help driftctl match IaC resources and remote resources
2) Filter noises from provider default resources
3) Resource transformation
4) Specific edge cases

## Examples

1) `aws_route_table_expander` explodes inline routes into dedicated resources
2) `aws_default_route` ignores routes created by default when creating a table **if they are not managed in IaC**
3) `aws_iam_user_policy_attachment` and `aws_iam_role_policy_attachment` transformed to `aws_iam_policy_attachment`
4) `route53_records_id_reconcilier` reworks IDs to match Terraform ones

> **Note:** As of v1.0.0, all non-AWS middlewares (Azure route/subnet expanders, Google IAM transformers) have been removed. Only AWS middlewares remain in `pkg/middlewares/`.
