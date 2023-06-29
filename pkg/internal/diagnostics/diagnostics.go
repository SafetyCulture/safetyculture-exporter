package diagnostics

import (
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// SysInfo contains high level system information that can be used to diagnose issues related to the underlying system.
type SysInfo struct {
	// Host
	OS                   string `json:"os"`              // ex: freebsd, linux
	Platform             string `json:"platform"`        // ex: ubuntu, linuxmint
	PlatformFamily       string `json:"platformFamily"`  // ex: debian, rhel
	PlatformVersion      string `json:"platformVersion"` // version of the complete OS
	KernelVersion        string `json:"kernelVersion"`   // version of the OS kernel (if available)
	KernelArch           string `json:"kernelArch"`      // native cpu architecture queried at runtime, as returned by `uname -m` or empty string in case of error
	VirtualizationSystem string `json:"virtualizationSystem"`
	VirtualizationRole   string `json:"virtualizationRole"` // guest or host

	// CPU
	CPUInfo []cpu.InfoStat

	// Memory
	// Total amount of RAM on this system
	MemoryTotal uint64 `json:"total"`

	// RAM available for programs to allocate
	//
	// This value is computed from the kernel specific values.
	MemoryAvailable uint64 `json:"available"`

	// RAM used by programs
	//
	// This value is computed from the kernel specific values.
	MemoryUsed uint64 `json:"used"`
}

// GetSysInfo returns high level system information that can be used to diagnose issues related to the underlying system.
func GetSysInfo() (*SysInfo, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get host info")
	}

	cpusInfo, err := cpu.Info()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get host info")
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get host info")
	}

	return &SysInfo{
		OS:                   hostInfo.OS,
		Platform:             hostInfo.Platform,
		PlatformFamily:       hostInfo.PlatformFamily,
		PlatformVersion:      hostInfo.PlatformVersion,
		KernelVersion:        hostInfo.KernelVersion,
		KernelArch:           hostInfo.KernelArch,
		VirtualizationSystem: hostInfo.VirtualizationSystem,
		VirtualizationRole:   hostInfo.VirtualizationRole,

		CPUInfo: cpusInfo,

		MemoryTotal:     memInfo.Total,
		MemoryAvailable: memInfo.Available,
		MemoryUsed:      memInfo.Used,
	}, nil

}
