package domains

import (
	"fmt"
)

var domainTypeIsEmpty = fmt.Errorf("Domain type is empty")
var domainIsEmpty = fmt.Errorf("Domain is empty")

func tooManyParts(d string) error {
	return fmt.Errorf(`Unable to parse. Domain "%s" has too many parts`, d)
}

func invalidDomain(d string) error {
	return fmt.Errorf(`Unable to parse. Domain "%s" is invalid`, d)
}

func invalidDomainType(t string) error {
	return fmt.Errorf(`Unable to parse. Domain type "%s" is invalid`, t)
}
