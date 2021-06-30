package paths

import "encoding/xml"

type SpaceChain []xml.Name

// IsExact determines if both chains are exactly equal
func (left SpaceChain) IsExact(right SpaceChain) bool {
	if len(left) != len(right) {
		return false
	}

	for i, leftItem := range left {
		rightItem := right[i]

		if leftItem != rightItem {
			return false
		}
	}

	return true
}

func makeSpaceChain(namespace string, locals ...string) SpaceChain {
	out := make(SpaceChain, len(locals))

	for i, local := range locals {
		name := xml.Name{
			Space: namespace,
			Local: local,
		}

		out[i] = name
	}

	return out
}
