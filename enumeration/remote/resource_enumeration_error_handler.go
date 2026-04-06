package remote

import (
	"errors"
	"net/http"
	"strings"

	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"

	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// HandleResourceEnumerationError inspects a resource enumeration error and raises an alert if appropriate.
func HandleResourceEnumerationError(err error, alerter alerter.Interface) error {
	var listError *remoteerror.ResourceScanningError
	if !errors.As(err, &listError) {
		return err
	}

	rootCause := listError.RootCause()

	// Check for AWS API errors (SDK v2)
	var respErr *smithyhttp.ResponseError
	if errors.As(rootCause, &respErr) {
		return handleAWSError(alerter, listError, respErr)
	}

	// Check for smithy API errors without HTTP response
	var apiErr smithy.APIError
	if errors.As(rootCause, &apiErr) {
		if strings.Contains(apiErr.ErrorCode(), "AccessDenied") {
			alerts.SendEnumerationAlert(common.RemoteAWSTerraform, alerter, listError)
			return nil
		}
	}

	// handles access denied errors in various message formats, e.g.:
	// aws_s3_bucket_policy: AccessDenied: Error listing bucket policy <policy_name>
	lowerMsg := strings.ToLower(rootCause.Error())
	if strings.Contains(lowerMsg, "accessdenied") || strings.Contains(lowerMsg, "access denied") {
		alerts.SendEnumerationAlert(common.RemoteAWSTerraform, alerter, listError)
		return nil
	}

	return err
}

func handleAWSError(alerter alerter.Interface, listError *remoteerror.ResourceScanningError, respErr *smithyhttp.ResponseError) error {
	statusCode := respErr.HTTPStatusCode()
	var apiErr smithy.APIError
	if errors.As(respErr, &apiErr) {
		if statusCode == http.StatusForbidden || (statusCode == http.StatusBadRequest && strings.Contains(apiErr.ErrorCode(), "AccessDenied")) {
			alerts.SendEnumerationAlert(common.RemoteAWSTerraform, alerter, listError)
			return nil
		}
	}

	return respErr
}
