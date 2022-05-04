package utils

import "sort"

// In Method like python in
func In(target string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, target)

	if index < len(strArray) && strArray[index] == target {
		return true
	}
	return false
}
