package hardware

import (
	"context"
	"sync"
	"time"

	"github.com/jaypipes/ghw"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/cluster/env/containerinfo"
)

const (
	usageSampleInterval = 200 * time.Millisecond
)

// hardwareCacheTTL bounds how often GetHardware() re-runs ghw's CPU/Memory/Block
// discovery (sysfs + block device enumeration), which is too costly to do on every
// call. It's a var, not a const, so tests can shrink it instead of sleeping 5s.
var hardwareCacheTTL = 5 * time.Second

type Manager struct {
	mu       sync.Mutex
	cached   *velez_api.GetHardware_Response
	cachedAt time.Time
}

func New() *Manager {
	return &Manager{}
}

func (h *Manager) GetHardware() (*velez_api.GetHardware_Response, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cached != nil && time.Since(h.cachedAt) < hardwareCacheTTL {
		return h.cached, nil
	}

	resp := &velez_api.GetHardware_Response{
		Cpu:                  &velez_api.GetHardware_Response_Value{},
		DiskMem:              &velez_api.GetHardware_Response_Value{},
		Ram:                  &velez_api.GetHardware_Response_Value{},
		IsRunningInContainer: containerinfo.IsInContainer(),
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

	h.cached = resp
	h.cachedAt = time.Now()

	return h.cached, nil
}

// GetUsage returns the current host-level CPU and memory utilization
// percentages of the machine this process runs on.
func (h *Manager) GetUsage(ctx context.Context) (cpuPercent, memPercent float64, err error) {
	cpuPercents, err := cpu.PercentWithContext(ctx, usageSampleInterval, false)
	if err != nil {
		return 0, 0, rerrors.Wrap(err, "error getting cpu usage")
	}

	if len(cpuPercents) > 0 {
		cpuPercent = cpuPercents[0]
	}

	virtualMemory, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return 0, 0, rerrors.Wrap(err, "error getting memory usage")
	}

	memPercent = virtualMemory.UsedPercent

	return cpuPercent, memPercent, nil
}
