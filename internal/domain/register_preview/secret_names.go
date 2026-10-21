package register_preview

import (
	"strings"
)

var secretNameMarkers = []string{"KEY", "TOKEN", "SECRET", "PASSWORD", "PASSWD", "CREDENTIAL", "AUTH"}

func isSecretName(name string) bool {
	upper := strings.ToUpper(name)

	for _, marker := range secretNameMarkers {
		if strings.Contains(upper, marker) {
			return true
		}
	}

	return false
}
