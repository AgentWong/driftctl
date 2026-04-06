package repository

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/snyk/driftctl/enumeration/remote/cache"
)

// CloudFormationRepository provides access to CloudFormation stack resource data.
type CloudFormationRepository interface {
	// ListAllStackResourcePhysicalIDs returns the set of physical resource IDs
	// managed by all active CloudFormation stacks in the account/region.
	ListAllStackResourcePhysicalIDs() (map[string]bool, error)
}

type cloudFormationRepository struct {
	client *cloudformation.Client
	cache  cache.Cache
}

// NewCloudFormationRepository creates a new CloudFormationRepository.
func NewCloudFormationRepository(cfg aws.Config, c cache.Cache) CloudFormationRepository {
	return &cloudFormationRepository{
		client: cloudformation.NewFromConfig(cfg),
		cache:  c,
	}
}

// ListAllStackResourcePhysicalIDs enumerates all active stacks and their resources,
// returning a set of physical resource IDs that CloudFormation manages.
func (r *cloudFormationRepository) ListAllStackResourcePhysicalIDs() (map[string]bool, error) {
	cacheKey := "cfnListAllStackResourcePhysicalIDs"
	v := r.cache.GetAndLock(cacheKey)
	defer r.cache.Unlock(cacheKey)
	if v != nil {
		return v.(map[string]bool), nil
	}

	physicalIDs := make(map[string]bool)

	// List all active stacks (exclude deleted)
	activeStatuses := []cftypes.StackStatus{
		cftypes.StackStatusCreateComplete,
		cftypes.StackStatusUpdateComplete,
		cftypes.StackStatusUpdateRollbackComplete,
		cftypes.StackStatusImportComplete,
		cftypes.StackStatusImportRollbackComplete,
		cftypes.StackStatusRollbackComplete,
	}

	stackPaginator := cloudformation.NewListStacksPaginator(r.client, &cloudformation.ListStacksInput{
		StackStatusFilter: activeStatuses,
	})
	for stackPaginator.HasMorePages() {
		page, err := stackPaginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, stack := range page.StackSummaries {
			stackName := aws.ToString(stack.StackName)
			if err := r.collectStackResources(stackName, physicalIDs); err != nil {
				return nil, err
			}
		}
	}

	r.cache.Put(cacheKey, physicalIDs)
	return physicalIDs, nil
}

// collectStackResources adds all physical resource IDs from a single stack.
func (r *cloudFormationRepository) collectStackResources(stackName string, ids map[string]bool) error {
	paginator := cloudformation.NewListStackResourcesPaginator(r.client, &cloudformation.ListStackResourcesInput{
		StackName: aws.String(stackName),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return err
		}
		for _, res := range page.StackResourceSummaries {
			if physID := aws.ToString(res.PhysicalResourceId); physID != "" {
				ids[physID] = true
			}
		}
	}
	return nil
}
