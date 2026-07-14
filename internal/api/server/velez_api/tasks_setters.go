package velez_api

// protoc-gen-go only emits Get* accessors, never Set*. These hand-written
// companions let jobs in internal/jobs mutate a task's persisted context
// through narrow accessor interfaces instead of a concrete struct type.

func (x *CreateSmerdTaskPayload) SetRequest(v *CreateSmerd_Request) {
	x.Request = v
}

func (x *CreateSmerdTaskPayload) SetImageId(v string) {
	x.ImageId = &v
}

func (x *CreateSmerdTaskPayload) SetContainerId(v string) {
	x.ContainerId = &v
}
