package experimental

import (
	"fmt"
	"regexp"
	"sort"
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

// ID pattern matches the JSON schema: letter, optionally followed by
// [letters|digits|_|-]* ending in alphanumeric. Single-character IDs are valid.
const idPattern = `[a-zA-Z](?:[a-zA-Z0-9_-]*[a-zA-Z0-9])?`

// regex
const (
	universePattern        = `^U:(` + idPattern + `)$`
	universeRealityPattern = `^U:(` + idPattern + `):(` + idPattern + `)$`
	realityPattern         = `^(` + idPattern + `)$`
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

func cloneStringSlice(src []string) []string {
	if src == nil {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func sortUniversesByID(universes []*ExUniverse) {
	sort.Slice(universes, func(i, j int) bool {
		return universes[i].model.ID < universes[j].model.ID
	})
}

func sortedMapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
