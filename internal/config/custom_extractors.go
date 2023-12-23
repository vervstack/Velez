package config

import (
	"strconv"
	"strings"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka"
)

func GetAvailablePorts(cfg Config) ([]uint16, error) {
	var portsSl []string
	err := matreshka.ReadSliceFromConfig(cfg.GetMatreshka(), AvailablePorts, &portsSl)
	if err != nil {
		return nil, err
	}

	ports := make([]uint16, 0, len(portsSl))
	for _, item := range portsSl {
		if idx := strings.Index(item, "-"); idx != -1 {
			portRange, err := parsePortRange(item[:idx], item[idx+1:])
			if err != nil {
				return nil, errors.Wrap(err, "error parsing ports range")
			}
			ports = append(ports, portRange...)
			continue
		}

		port, err := strconv.ParseUint(item, 10, 16)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing value")
		}
		ports = append(ports, uint16(port))
	}

	return ports, nil
}

func parsePortRange(start, end string) ([]uint16, error) {
	portFrom64, err := strconv.ParseUint(start, 10, 16)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing value")
	}

	portTo64, err := strconv.ParseUint(end, 10, 16)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing value")
	}

	portTo := uint16(portTo64)
	portFrom := uint16(portFrom64)
	if portTo < portFrom {
		portFrom = portTo
	}

	ports := make([]uint16, 0, portTo-portFrom+1)

	for i := portFrom; i <= portTo; i++ {
		ports = append(ports, i)
	}

	return ports, nil
}
