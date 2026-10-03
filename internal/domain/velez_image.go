package domain

import (
	"strings"
)

const (
	velezImageMarker = "velez"
)

// IsVelezImage reports whether image is a Velez image: its repository, with
// registry host and path included but tag and digest dropped, contains "velez".
func IsVelezImage(image string) bool {
	repository, _, _ := strings.Cut(image, "@")

	lastSlash := strings.LastIndex(repository, "/")

	lastColon := strings.LastIndex(repository, ":")
	if lastColon > lastSlash {
		repository = repository[:lastColon]
	}

	return strings.Contains(strings.ToLower(repository), velezImageMarker)
}
