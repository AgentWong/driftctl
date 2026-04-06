package enumeration

// ProgressCounter tracks the number of resources processed.
type ProgressCounter interface {
	Inc()
}
