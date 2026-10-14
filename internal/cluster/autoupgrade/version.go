package autoupgrade

import (
	"regexp"
	"strings"

	"golang.org/x/mod/semver"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

var strictSemverPattern = regexp.MustCompile(`^v?\d+\.\d+\.\d+$`)

type imageRef struct {
	Repo string
	Tag  string
}

func splitImage(image string) imageRef {
	colonIndex := strings.LastIndex(image, ":")
	slashIndex := strings.LastIndex(image, "/")

	if colonIndex == -1 || colonIndex < slashIndex {
		return imageRef{Repo: image}
	}

	return imageRef{
		Repo: image[:colonIndex],
		Tag:  image[colonIndex+1:],
	}
}

// pickTag returns the tag to move to and whether to move at all.
func pickTag(currentTag string, tags []string, level velez_api.AutoUpdateBumpLevel) (string, bool) {
	if !isStrictSemver(currentTag) {
		return pickLegacyTag(currentTag, tags)
	}

	current := toSemver(currentTag)

	bestTag := ""
	best := ""

	for _, tag := range tags {
		if !isStrictSemver(tag) {
			continue
		}

		candidate := toSemver(tag)
		if semver.Compare(candidate, current) <= 0 {
			continue
		}

		if !isBumpAccepted(current, candidate, level) {
			continue
		}

		if best == "" || semver.Compare(candidate, best) > 0 {
			best = candidate
			bestTag = tag
		}
	}

	return bestTag, bestTag != ""
}

func pickLegacyTag(currentTag string, tags []string) (string, bool) {
	if len(tags) == 0 {
		return "", false
	}

	last := tags[len(tags)-1]
	if last == currentTag {
		return "", false
	}

	return last, true
}

func isBumpAccepted(current, candidate string, level velez_api.AutoUpdateBumpLevel) bool {
	isMajorBump := semver.Major(candidate) != semver.Major(current)
	if isMajorBump {
		return true
	}

	isMinorBump := semver.MajorMinor(candidate) != semver.MajorMinor(current)

	switch level {
	case velez_api.AutoUpdateBumpLevel_MAJOR:
		return false
	case velez_api.AutoUpdateBumpLevel_MINOR:
		return isMinorBump
	default:
		return true
	}
}

func isStrictSemver(tag string) bool {
	return strictSemverPattern.MatchString(tag)
}

func toSemver(tag string) string {
	if strings.HasPrefix(tag, "v") {
		return tag
	}

	return "v" + tag
}
