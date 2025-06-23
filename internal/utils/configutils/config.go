package configutils

import (
	"strings"

	"go.vervstack.ru/matreshka/pkg/matreshka_api"
)

func AppendPrefix(prefix matreshka_api.ConfigTypePrefix, name string) string {
	if prefix == matreshka_api.ConfigTypePrefix_plain {
		return name
	}

	if !strings.HasPrefix(name, prefix.String()) {
		name = prefix.String() + "_" + name
	}

	return name
}
