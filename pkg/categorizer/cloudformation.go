package categorizer

import (
	"regexp"
	"strings"

	"github.com/snyk/driftctl/enumeration/resource"
)

// cfnPhysicalIDPattern matches CloudFormation-generated physical resource IDs,
// which follow the format <stack>-<LogicalId>-<12-or-13-char random suffix>.
var cfnPhysicalIDPattern = regexp.MustCompile(`^.+-.+-[A-Za-z0-9]{12,13}$`)

// cdkNamePattern matches resource names from CDK and AWS Solutions stacks that
// use <lowercase-prefix>-<CamelCaseLogicalId>[-<CamelCaseSuffix>] naming.
var cdkNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[A-Z][a-zA-Z0-9]*)+$`)

// uuidPattern matches UUID-formatted strings (e.g. KMS key IDs) that should
// not be confused with CloudFormation physical IDs.
var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// CloudFormationCategorizer detects resources managed by CloudFormation stacks.
// It first checks for the aws:cloudformation:stack-name tag, then falls back to
// matching the CloudFormation physical-ID naming convention.
type CloudFormationCategorizer struct{}

// NewCloudFormationCategorizer creates a new CloudFormationCategorizer.
func NewCloudFormationCategorizer() *CloudFormationCategorizer {
	return &CloudFormationCategorizer{}
}

// Categorize returns CategoryCloudFormationManaged if the resource appears to be managed by CloudFormation.
func (c *CloudFormationCategorizer) Categorize(r *resource.Resource) (Category, bool) {
	if matchesCfnTag(r) {
		return CategoryCloudFormationManaged, true
	}
	if matchesCfnNamePattern(r) {
		return CategoryCloudFormationManaged, true
	}
	return "", false
}

// matchesCfnTag checks for the aws:cloudformation:stack-name tag.
func matchesCfnTag(r *resource.Resource) bool {
	attrs := r.Attributes()
	if attrs == nil {
		return false
	}

	tags, ok := (*attrs)["tags"]
	if !ok {
		return false
	}

	switch t := tags.(type) {
	case map[string]interface{}:
		if _, hasStackTag := t["aws:cloudformation:stack-name"]; hasStackTag {
			return true
		}
	case map[string]string:
		if _, hasStackTag := t["aws:cloudformation:stack-name"]; hasStackTag {
			return true
		}
	}

	return false
}

// matchesCfnNamePattern detects resources whose name or ID matches the
// CloudFormation physical-ID convention: <stack>-<LogicalId>-<12-char suffix>.
func matchesCfnNamePattern(r *resource.Resource) bool {
	name := resourceName(r)
	id := r.ResourceId()

	// "AwsSolutions" path indicates an AWS Solutions CloudFormation stack
	if strings.Contains(name, "/AwsSolutions/") || strings.Contains(id, "/AwsSolutions/") {
		return true
	}

	if name != "" && !looksLikeUUIDOrARN(name) {
		if cfnPhysicalIDPattern.MatchString(name) || cdkNamePattern.MatchString(name) {
			return true
		}
	}

	// for SQS queues, the ID is a URL — extract the queue name from the path
	if r.ResourceType() == "aws_sqs_queue" {
		if idx := lastSlashIndex(id); idx >= 0 {
			id = id[idx+1:]
		}
	}

	// UUIDs and ARNs contain segments that look like CFn suffixes but aren't
	if looksLikeUUIDOrARN(id) {
		return false
	}

	return cfnPhysicalIDPattern.MatchString(id) || cdkNamePattern.MatchString(id)
}

// looksLikeUUIDOrARN returns true when s is in ARN or UUID format, both of
// which contain 12-char hex segments that would false-positive the CFn regex.
func looksLikeUUIDOrARN(s string) bool {
	if strings.HasPrefix(s, "arn:") {
		return true
	}
	return uuidPattern.MatchString(s)
}

func lastSlashIndex(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			return i
		}
	}
	return -1
}
