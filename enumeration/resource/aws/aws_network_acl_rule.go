package aws

import (
	"bytes"
	"fmt"

	"github.com/snyk/driftctl/pkg/helpers"
)

// AwsNetworkACLRuleResourceType is the resource type for AWS Network ACL rules.
const AwsNetworkACLRuleResourceType = "aws_network_acl_rule"

// CreateNetworkACLRuleID creates a unique identifier for a network ACL rule based on its properties.
func CreateNetworkACLRuleID(networkACLID string, ruleNumber int64, egress bool, protocol string) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s-", networkACLID)
	fmt.Fprintf(&buf, "%d-", ruleNumber)
	fmt.Fprintf(&buf, "%t-", egress)
	fmt.Fprintf(&buf, "%s-", protocol)
	return fmt.Sprintf("nacl-%d", helpers.HashcodeString(buf.String()))
}
