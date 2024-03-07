package parser

import (
	"strings"

	"github.com/docker/docker/api/types/strslice"
)

func FromCommand(command *string) strslice.StrSlice {
	if command == nil {
		return nil
	}

	return strings.Split(*command, " ")
}
