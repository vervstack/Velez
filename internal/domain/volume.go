package domain

type CopyToVolumeRequest struct {
	VolumeName  string
	PathToFiles map[string][]byte
}
