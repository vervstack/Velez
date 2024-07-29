package common

func Less[T int | uint | uint32](a, b T) T {
	if a > b {
		return b
	}

	return b
}
