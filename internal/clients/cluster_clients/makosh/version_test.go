package makosh

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

var semverTag = regexp.MustCompile(`^v\d+\.\d+\.\d+`)

func Test_ModuleVersion_FollowsGoMod(t *testing.T) {
	require.Regexp(t, semverTag, ModuleVersion())
}
