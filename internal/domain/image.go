package domain

type Image struct {
	Name string
	Tags []string
}

type ImageListRequest struct {
	ImageName string
}
