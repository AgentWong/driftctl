package resource

// Supplier supply the list of resource.Resource, it's the main interface to retrieve remote resources
type Supplier interface {
	Resources() ([]*Resource, error)
}

// StoppableSupplier extends Supplier with a Stop method for cleanup.
type StoppableSupplier interface {
	Supplier
	Stop()
}
