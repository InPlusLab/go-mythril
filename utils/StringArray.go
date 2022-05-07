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

// Zfill
func Zfill(str string, size int) string {
	res := str
	strLen := len(str)
	if strLen < size {
		for i := 0; i < size-strLen; i++ {
			res = "0" + res
		}
	}
	return res
}
