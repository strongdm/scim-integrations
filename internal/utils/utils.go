package utils

func StringArrayContains(array []string, x string) bool {
	for _, item := range array {
		if item == x {
			return true
		}
	}
	return false
}
