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

func (x *AssembleConfigTaskPayload) SetImageLabels(v map[string]string) {
	x.ImageLabels = v
}

func (x *AssembleConfigTaskPayload) SetImageTags(v []string) {
	x.ImageTags = v
}

func (x *AssembleConfigTaskPayload) SetContainerId(v string) {
	x.ContainerId = &v
}

func (x *AssembleConfigTaskPayload) SetConfigName(v string) {
	x.ConfigName = &v
}

func (x *AssembleConfigTaskPayload) SetConfigVersion(v string) {
	x.ConfigVersion = &v
}

func (x *AssembleConfigTaskPayload) SetConfType(v string) {
	x.ConfType = &v
}

func (x *AssembleConfigTaskPayload) SetConfigFormat(v ConfigFormat) {
	x.ConfigFormat = &v
}

func (x *AssembleConfigTaskPayload) SetContentRaw(v []byte) {
	x.ContentRaw = v
}

func (x *AssembleConfigTaskPayload) SetContent(v []byte) {
	x.Content = v
}
