package utils

func FindIndex(arr []string, value string) int {
	for k, v := range arr {
		if v == value {
			return k
		}
	}
	return -1
}
