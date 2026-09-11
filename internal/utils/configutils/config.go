package configutils

import (
	"strings"

	"go.vervstack.ru/Velez/internal/api/clients/matreshka/pkg/matreshka_api"
)

func AppendPrefix(prefix matreshka_api.ConfigType, name string) string {
	if prefix == matreshka_api.ConfigType_plain {
		return name
	}

	if !strings.HasPrefix(name, prefix.String()) {
		name = prefix.String() + "_" + name
	}

	return name
}
