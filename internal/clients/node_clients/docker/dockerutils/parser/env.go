package parser

import (
	"strings"
)

func FromDockerEnv(env map[string]string) []string {
	out := make([]string, 0, len(env))

	for k, v := range env {
		out = append(out, k+"="+v)
	}

	return out
}

func ToDockerEnv(env []string) map[string]string {
	out := make(map[string]string, len(env))

	for _, e := range env {
		nameVal := strings.Split(e, "=")
		if len(nameVal) < 2 {
			continue
		}

		out[nameVal[0]] = strings.Join(nameVal[1:], "")
	}

	return out
}
