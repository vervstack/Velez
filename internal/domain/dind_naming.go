package domain

const (
	dindDaemonPort     = "2375"
	dindNetworkSuffix  = "-net"
	DindImage          = "docker:dind"
	DindDataVolumePath = "/var/lib/docker"
)

func DindAddress(name string) string {
	return "tcp://" + name + ":" + dindDaemonPort
}

func DindNetworkName(name string) string {
	return name + dindNetworkSuffix
}

func DindDataVolumeName(name string) string {
	return name + "-data"
}
