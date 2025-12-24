package domain

type EnableStatefullClusterRequest struct {
	ExposePort   bool
	ExposeToPort uint64
}
