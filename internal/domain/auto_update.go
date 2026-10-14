package domain

import (
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	bumpLevelLabelMajor = "major"
	bumpLevelLabelMinor = "minor"
	bumpLevelLabelPatch = "patch"
)

// BumpLevelFromLabels reads the auto-update bump level off container labels;
// absent or unknown values mean patch.
func BumpLevelFromLabels(containerLabels map[string]string) velez_api.AutoUpdateBumpLevel {
	switch containerLabels[labels.AutoUpgradeBumpLevel] {
	case bumpLevelLabelMajor:
		return velez_api.AutoUpdateBumpLevel_MAJOR
	case bumpLevelLabelMinor:
		return velez_api.AutoUpdateBumpLevel_MINOR
	default:
		return velez_api.AutoUpdateBumpLevel_PATCH
	}
}

// BumpLevelLabelValue is the inverse of BumpLevelFromLabels; an unspecified
// level is written as patch.
func BumpLevelLabelValue(level velez_api.AutoUpdateBumpLevel) string {
	switch level {
	case velez_api.AutoUpdateBumpLevel_MAJOR:
		return bumpLevelLabelMajor
	case velez_api.AutoUpdateBumpLevel_MINOR:
		return bumpLevelLabelMinor
	default:
		return bumpLevelLabelPatch
	}
}
