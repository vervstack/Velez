package jobs

import (
	"github.com/distribution/reference"
	"go.redsock.ru/rerrors"
)

// imageWithTag returns image with its tag replaced by tag; a digest reference loses its digest.
func imageWithTag(image, tag string) (string, error) {
	named, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return "", rerrors.Wrap(err, "error parsing image reference")
	}

	trimmed := reference.TrimNamed(named)

	tagged, err := reference.WithTag(trimmed, tag)
	if err != nil {
		return "", rerrors.Wrap(err, "error tagging image reference")
	}

	return reference.FamiliarString(tagged), nil
}
