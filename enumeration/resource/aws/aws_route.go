package aws

import (
	"fmt"

	"github.com/snyk/driftctl/pkg/helpers"
)

const AwsRouteResourceType = "aws_route"

func CalculateRouteID(tableId, CidrBlock, Ipv6CidrBlock, PrefixListId *string) string {
	if CidrBlock != nil && *CidrBlock != "" {
		return fmt.Sprintf("r-%s%d", *tableId, helpers.HashcodeString(*CidrBlock))
	}

	if Ipv6CidrBlock != nil && *Ipv6CidrBlock != "" {
		return fmt.Sprintf("r-%s%d", *tableId, helpers.HashcodeString(*Ipv6CidrBlock))
	}

	if PrefixListId != nil && *PrefixListId != "" {
		return fmt.Sprintf("r-%s%d", *tableId, helpers.HashcodeString(*PrefixListId))
	}

	return ""
}
