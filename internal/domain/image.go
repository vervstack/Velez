package domain

type ImageListRequest struct {
	Name string
}

type ImageSearchRequest struct {
	Term            string
	UseOfficialOnly bool
}
