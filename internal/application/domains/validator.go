package domains

import "regexp"

var domainRegex = regexp.MustCompile(`^((?:(?!-)[A-Za-z0-9-]{1,63}(?<!-)\.)+[A-Za-z]{2,})$`)

func isValid(domain string) bool {
	return domainRegex.MatchString(domain)
}
