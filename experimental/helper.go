package experimental

import (
	"fmt"
	"regexp"
)

type refType int

// <universeId> and <realityId> validations:
//	* no white space
//	* only letters, numbers, underscore (_) and dash (-)
// 	* must start with a letter
//	* min length: 1

const (
	// RefTypeUniverse format -> U:<universeId>
	RefTypeUniverse refType = iota

	// RefTypeUniverseReality format -> U:<universeId>:<realityId>
	RefTypeUniverseReality

	// RefTypeReality format -> <realityId>
	RefTypeReality
)

// regex
const (
	universePattern        = `^U:([a-zA-Z][a-zA-Z0-9_-]*[a-zA-Z0-9])$`
	universeRealityPattern = `^U:([a-zA-Z][a-zA-Z0-9_-]*[a-zA-Z0-9]):([a-zA-Z][a-zA-Z0-9_-]*[a-zA-Z0-9])$`
	realityPattern         = `^([a-zA-Z][a-zA-Z0-9_-]*[a-zA-Z0-9])$`
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

// processReference returns the refType and parts of the ref
func processReference(ref string) (refType, []string, error) {
	for _, p := range patterns {
		matches := p.regEx.FindStringSubmatch(ref)
		if matches != nil {
			// Remove the full string match and return only capturing groups
			return p.ref, matches[1:], nil
		}
	}
	return -1, nil, fmt.Errorf("invalid ref '%s'", ref)
}
