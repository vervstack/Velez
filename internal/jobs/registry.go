package jobs

// Registry maps a task's action name to the TaskHandler that knows how to
// build its Job list.
type Registry struct {
	handlers map[string]TaskHandler
}

func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]TaskHandler),
	}
}

func (r *Registry) Register(h TaskHandler) {
	r.handlers[h.Action()] = h
}

func (r *Registry) Get(action string) (TaskHandler, bool) {
	h, ok := r.handlers[action]

	return h, ok
}
