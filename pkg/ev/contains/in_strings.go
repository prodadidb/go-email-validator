package contains

import "github.com/emirpasic/gods/sets/hashset"

// InStrings checks existing value in set of strings
type InStrings interface {
	Contains(value string) bool
}

// NewInStringsFromArray instantiates InStrings based on []string
func NewInStringsFromArray(elements []string) InStrings {
	var maxLen = 0
	setElements := make([]interface{}, len(elements))
	for i, element := range elements {
		currentLen := len(element)
		if currentLen > maxLen {
			maxLen = currentLen
		}
		setElements[i] = element
	}

	return NewInStrings(NewSet(hashset.New(setElements...)), maxLen)
}

// NewInStrings instantiates InStrings based on InSet
func NewInStrings(contains InSet, maxLen int) InStrings {
	return InStringsStructs{contains, maxLen}
}

type InStringsStructs struct {
	ContainsInSet InSet
	MaxLen        int
}

func (is InStringsStructs) Contains(value string) bool {
	var maxLen = len(value)
	var jEnd int

	for i := 0; i < maxLen; i++ {
		key := ""
		jEnd = i + is.MaxLen
		if jEnd > maxLen {
			jEnd = maxLen
		}
		for j := i; j < jEnd; j++ {
			key += string(value[j])
			if is.ContainsInSet.Contains(key) {
				return true
			}
		}
	}

	return false
}
