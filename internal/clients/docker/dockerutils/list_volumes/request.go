package list_volumes

import (
	"github.com/docker/docker/api/types/filters"
)

type Filter struct {
	args filters.Args
}

func New() Filter {
	return Filter{filters.NewArgs()}
}

func (f *Filter) Name(name string) {
	f.args.Add("name", name)
}

func (f *Filter) Args() filters.Args {
	return f.args
}
