package node_clients

import (
	"net/url"
)

const (
	goosDarwin  = "darwin"
	goosLinux   = "linux"
	goosWindows = "windows"

	pingHintPrefix = "can't ping docker api at "

	pingHintInContainer = "mount the docker socket into the Velez container: " +
		"-v /var/run/docker.sock:/var/run/docker.sock"
	pingHintRemoteDaemon = "make sure dockerd on that host listens on this address " +
		"(\"hosts\" in /etc/docker/daemon.json) and the port is reachable from this machine"
	pingHintDarwinLocalNetwork = "macOS blocks LAN connections for apps without Local Network access " +
		"(\"connect: no route to host\" while curl works): System Settings > Privacy & Security > " +
		"Local Network, enable the terminal or IDE that runs Velez and restart it"
	pingHintWindowsFirewall = "allow outbound connections for Velez in Windows Defender Firewall"
	pingHintDarwinLocal     = "start Docker Desktop (or colima / OrbStack) and point DOCKER_HOST " +
		"or the active docker context at its socket"
	pingHintLinuxLocal = "start dockerd (systemctl start docker) and make sure the user running Velez " +
		"can access /var/run/docker.sock (docker group)"
	pingHintWindowsLocal = "start Docker Desktop and make sure the npipe:////./pipe/docker_engine pipe exists"
)

func dockerPingHint(goos, daemonHost string, isInContainer bool) string {
	hint := pingHintPrefix + daemonHost + ": "

	if isRemoteDaemon(daemonHost) {
		hint += pingHintRemoteDaemon

		switch goos {
		case goosDarwin:
			return hint + "; " + pingHintDarwinLocalNetwork
		case goosWindows:
			return hint + "; " + pingHintWindowsFirewall
		}

		return hint
	}

	if isInContainer {
		return hint + pingHintInContainer
	}

	switch goos {
	case goosDarwin:
		return hint + pingHintDarwinLocal
	case goosLinux:
		return hint + pingHintLinuxLocal
	case goosWindows:
		return hint + pingHintWindowsLocal
	}

	return hint + pingHintLinuxLocal
}

func isRemoteDaemon(daemonHost string) bool {
	parsed, err := url.Parse(daemonHost)
	if err != nil {
		return false
	}

	switch parsed.Scheme {
	case "tcp", "http", "https":
		return true
	}

	return false
}
