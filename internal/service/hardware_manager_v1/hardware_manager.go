package hardware_manager_v1

import (
	"github.com/jaypipes/ghw"

	"github.com/godverv/Velez/pkg/velez_api"
)

type HardwareManager struct {
}

func New() *HardwareManager {
	return &HardwareManager{}
}

func (h *HardwareManager) GetHardware() (*velez_api.GetHardware_Response, error) {
	resp := &velez_api.GetHardware_Response{
		Cpu:     &velez_api.GetHardware_Response_Value{},
		DiskMem: &velez_api.GetHardware_Response_Value{},
		Ram:     &velez_api.GetHardware_Response_Value{},
	}

	cpu, err := ghw.CPU()
	if err != nil {
		resp.Cpu.Err = err.Error()
	} else {
		resp.Cpu.Value = cpu.String()
	}

	ram, err := ghw.Memory()
	if err != nil {
		resp.Ram.Err = err.Error()
	} else {
		resp.Ram.Value = ram.String()
	}

	block, err := ghw.Block()
	if err != nil {
		resp.DiskMem.Err = err.Error()
	} else {
		resp.DiskMem.Value = block.String()
	}

	return resp, nil
}
