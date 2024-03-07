package list_request

import (
	"strconv"

	"github.com/docker/docker/api/types/filters"
)

type Filter struct {
	args filters.Args
}

func New() Filter {
	return Filter{filters.NewArgs()}
}

// Exited add filter with for containers with exit code of exitCode
func (f *Filter) Exited(exitCode int) {
	f.args.Add("exited", strconv.Itoa(exitCode))
}

type status string

const (
	Created    status = "created"
	Restarting status = "restarting"
	Running    status = "running"
	Removing   status = "removing"
	Paused     status = "paused"
	Exited     status = "exited"
	Dead       status = "dead"
)

func (f *Filter) Status(in status) {
	f.args.Add("status", string(in))
}

func (f *Filter) Label(label string) {
	f.args.Add("label", label)
}

func (f *Filter) Id(id string) {
	f.args.Add("id", id)
}

func (f *Filter) Name(name string) {
	f.args.Add("name", name)
}

func (f *Filter) IsTask(isTask bool) {
	f.args.Add("is-task", strconv.FormatBool(isTask))
}

type health string

const (
	Starting  health = "starting"
	Healthy   health = "healthy"
	Unhealthy health = "unhealthy"
	None      health = "none"
)

func (f *Filter) Health(h health) {
	f.args.Add("health", string(h))
}

func (f *Filter) Args() filters.Args {
	return f.args.Clone()
}
