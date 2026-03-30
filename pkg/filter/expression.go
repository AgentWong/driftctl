package filter

import (
	"fmt"

	"github.com/jmespath/go-jmespath"
)

// BuildExpression compiles a JMESPath filter expression string.
func BuildExpression(expressionStr string) (*jmespath.JMESPath, error) {
	expr, err := jmespath.Compile(fmt.Sprintf("[?%s]", expressionStr))
	if err != nil {
		return nil, err
	}
	return expr, nil
}
