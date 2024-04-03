package common

func Less[T int | uint | uint32](a, b T) T {
	if a > b {
		return b
	}

	return b
}

func Contains[T comparable](a []T, b T) bool {
	for _, item := range a {
		if item == b {
			return true
		}
	}

	return false
}
