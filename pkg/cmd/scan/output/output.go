package output

import (
	"sort"

	"github.com/snyk/driftctl/pkg/analyser"
	"github.com/snyk/driftctl/pkg/output"
)

// Output is the interface implemented by all scan output writers.
type Output interface {
	Write(analysis *analyser.Analysis) error
}

var supportedOutputTypes = []string{
	ConsoleOutputType,
	JSONOutputType,
	HTMLOutputType,
	PlanOutputType,
}

var supportedOutputExample = map[string]string{
	ConsoleOutputType: ConsoleOutputExample,
	JSONOutputType:    JSONOutputExample,
	HTMLOutputType:    HTMLOutputExample,
	PlanOutputType:    PlanOutputExample,
}

// SupportedOutputsExample returns a sorted list of example URIs for all supported output types.
func SupportedOutputsExample() []string {
	examples := make([]string, 0, len(supportedOutputExample))
	for _, ex := range supportedOutputExample {
		examples = append(examples, ex)
	}
	sort.Strings(examples)
	return examples
}

// Example returns the example URI string for the given output type key.
func Example(key string) string {
	return supportedOutputExample[key]
}

// IsSupported reports whether the given key corresponds to a supported output type.
func IsSupported(key string) bool {
	for _, o := range supportedOutputTypes {
		if o == key {
			return true
		}
	}
	return false
}

// GetOutput returns the Output implementation for the given Config.
func GetOutput(config Config) Output {
	switch config.Key {
	case JSONOutputType:
		return NewJSON(config.Path)
	case HTMLOutputType:
		return NewHTML(config.Path)
	case PlanOutputType:
		return NewPlan(config.Path)
	case ConsoleOutputType:
		fallthrough
	default:
		return NewConsole()
	}
}

// ShouldPrint indicate if we should use the global output or not (e.g. when outputting to stdout).
func ShouldPrint(outputs []Config, quiet bool) bool {
	for _, c := range outputs {
		p := GetPrinter(c, quiet)
		if _, ok := p.(*output.VoidPrinter); ok {
			return false
		}
	}
	return true
}

// GetPrinter returns the appropriate Printer for the given Config and quiet flag.
func GetPrinter(config Config, quiet bool) output.Printer {
	if quiet {
		return &output.VoidPrinter{}
	}

	switch config.Key {
	case JSONOutputType:
		fallthrough
	case PlanOutputType:
		fallthrough
	case HTMLOutputType:
		fallthrough
	case ConsoleOutputType:
		fallthrough
	default:
		return output.NewConsolePrinter()
	}
}

func isStdOut(path string) bool {
	return path == "/dev/stdout" || path == "stdout"
}
