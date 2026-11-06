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

// Actions lists every registered action name.
func (r *Registry) Actions() []string {
	actions := make([]string, 0, len(r.handlers))
	for action := range r.handlers {
		actions = append(actions, action)
	}

	return actions
}
