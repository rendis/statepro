package experimental

import (
	"fmt"
	"regexp"
)

type refType int

const (
	// RefTypeUniverse format: <universeId>@<universeVersion>
	RefTypeUniverse refType = iota

	// RefTypeUniverseReality format: <universeId>@<universeVersion>:<realityId>
	RefTypeUniverseReality

	// RefTypeReality format: <realityId>
	RefTypeReality
)

// regex
const (
	universePattern        = `^([^@]{1,}@[^@]{1,})$`
	universeRealityPattern = `^([^@]{1,}@[^@]{1,}):([^:]{1,})$`
	realityPattern         = `^([^@:]{1,})$`
)

type refTypePattern struct {
	ref   refType
	regEx *regexp.Regexp
}

// patterns establishes the order of evaluation of the regex
var patterns = []refTypePattern{
	{ref: RefTypeUniverseReality, regEx: regexp.MustCompile(universeRealityPattern)},
	{ref: RefTypeUniverse, regEx: regexp.MustCompile(universePattern)},
	{ref: RefTypeReality, regEx: regexp.MustCompile(realityPattern)},
}

// getReferenceType returns the refType and parts of the ref
func getReferenceType(ref string) (refType, []string, error) {
	for _, p := range patterns {
		matches := p.regEx.FindStringSubmatch(ref)
		if matches != nil {
			// Remove the full string match and return only capturing groups
			return p.ref, matches[1:], nil
		}
	}
	return -1, nil, fmt.Errorf("invalid ref '%s'", ref)
}
