// Package awsutils provides shared AWS session helpers for acceptance tests.
package awsutils

import (
	"github.com/aws/aws-sdk-go/aws/session"
)

// Session creates a new AWS session using the shared config (e.g. ~/.aws/config).
func Session() *session.Session {
	return session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
}
