package utils

import (
	"strings"
)

// IsSubsequence takes two strings and returns true
// if the second string is a subsequence of the first.
// If ignoreCase is set, the comparison will be
// case insensitive.
// https://en.wikipedia.org/wiki/Subsequence
func IsSubsequence(input, target string, ignoreCase bool) bool {
	if ignoreCase {
		input = strings.ToLower(input)
		target = strings.ToLower(target)
	}

	idx := 0
	targetArray := []rune(target)

	for _, r := range input {
		if len(targetArray) > idx && targetArray[idx] == r {
			idx++
		}
	}

	return idx == len(targetArray)
}
