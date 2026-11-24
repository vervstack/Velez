package domain

import (
	"time"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	HaproxyDefaultPort = 5432
	HaproxyMaxPort     = 65535
	HaproxyConfigPath  = "/usr/local/etc/haproxy/haproxy.cfg"
)

type HaproxyInstance struct {
	Name             string
	MasterPgInstance string
	Port             uint32
	Environment      string
	Status           string
	CreatedAt        time.Time
}

func HaproxyServiceName(name string) string {
	return labels.HaproxyNamePrefix + name
}

func HaproxyConfigVolumeName(name string) string {
	return HaproxyServiceName(name) + "-config"
}
