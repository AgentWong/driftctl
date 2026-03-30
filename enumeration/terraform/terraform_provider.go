package terraform

// Provider defines the interface for a Terraform provider that supplies schemas and reads resources.
type Provider interface {
	SchemaSupplier
	ResourceReader
	Cleanup()
	Name() string
	Version() string
}
