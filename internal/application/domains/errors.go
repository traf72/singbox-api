package domains

import (
	"fmt"
)

var errEmptyTemplate = fmt.Errorf("template is empty")
var errEmptyTemplateType = fmt.Errorf("template type is empty")

func tooManyParts(d string) error {
	return fmt.Errorf(`template "%s" has too many parts`, d)
}

func invalidDomain(d string) error {
	return fmt.Errorf(`domain "%s" is invalid`, d)
}

func invalidTemplateType(t string) error {
	return fmt.Errorf(`template type "%s" is invalid`, t)
}
