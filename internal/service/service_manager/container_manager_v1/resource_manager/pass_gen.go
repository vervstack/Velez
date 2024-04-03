package resource_manager

import (
	"math/rand"
)

func genPass() string {
	const (
		digits = "0123456789"
		all    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
			"abcdefghijklmnopqrstuvwxyz" +
			digits
	)

	length := 8

	buf := make([]byte, length)
	buf[0] = digits[rand.Intn(len(digits))]

	for i := 1; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}

	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})

	return string(buf)
}
