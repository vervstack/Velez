package domain

type ContainerCreate struct {
	ImageName string
}

type Container struct {
	ImageName string
	Tags      []string
}
