package aws

import (
	"bytes"
	"fmt"

	"github.com/snyk/driftctl/pkg/helpers"
)

const AwsNetworkACLRuleResourceType = "aws_network_acl_rule"

func CreateNetworkACLRuleID(networkAclId string, ruleNumber int64, egress bool, protocol string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", networkAclId))
	buf.WriteString(fmt.Sprintf("%d-", ruleNumber))
	buf.WriteString(fmt.Sprintf("%t-", egress))
	buf.WriteString(fmt.Sprintf("%s-", protocol))
	return fmt.Sprintf("nacl-%d", helpers.HashcodeString(buf.String()))
}
