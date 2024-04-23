package list_request

import (
	"github.com/docker/docker/api/types/filters"
)

type Filter struct {
	args filters.Args
}

func New() Filter {
	return Filter{filters.NewArgs()}
}
