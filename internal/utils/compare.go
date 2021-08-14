package utils

func CompareSliceString(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	for key, val := range a {
		if val != b[key] {
			return false
		}
	}

	return true
}
