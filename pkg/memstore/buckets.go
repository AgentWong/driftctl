package memstore

// BucketName identifies a named bucket in the store.
type BucketName int

const (
	// TelemetryBucket is the name of the store bucket used by the telemetry service
	TelemetryBucket BucketName = iota + 1
)
