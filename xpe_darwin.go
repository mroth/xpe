package xpe

import (
	"fmt"
	"syscall"
)

type sysctler interface {
	Sysctl(name string) (value string, err error)
	SysctlUint32(name string) (value uint32, err error)
}

type nativeSyscallPackage struct{}

func (s nativeSyscallPackage) Sysctl(name string) (value string, err error) {
	return syscall.Sysctl(name)
}

func (s nativeSyscallPackage) SysctlUint32(name string) (value uint32, err error) {
	return syscall.SysctlUint32(name)
}

var darwinsys sysctler = nativeSyscallPackage{}

func GetCPU() (*CPU, error) {
	const (
		cpuBrandString = "machdep.cpu.brand_string"
		cpuThreadCount = "machdep.cpu.thread_count"
		cpuCoreCount   = "machdep.cpu.core_count"
		hwNPerflevels  = "hw.nperflevels"
		lpKey          = "hw.perflevel0.logicalcpu"
		leKey          = "hw.perflevel1.logicalcpu"
		ppKey          = "hw.perflevel0.physicalcpu"
		peKey          = "hw.perflevel1.physicalcpu"
	)

	brandString, err := darwinsys.Sysctl(cpuBrandString)
	if err != nil {
		return nil, fmt.Errorf("failed to get %v: %w", cpuBrandString, err)
	}

	threads, err := darwinsys.SysctlUint32(cpuThreadCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get %v: %w", cpuThreadCount, err)
	}

	cores, err := darwinsys.SysctlUint32(cpuCoreCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get %v: %w", cpuCoreCount, err)
	}

	perfLevels, err := darwinsys.SysctlUint32(hwNPerflevels)
	if err != nil {
		// TODO: seems might not exist on older versions of macOS? lets not error if so
		return &CPU{
			BrandString: brandString,
			Threads:     int(threads),
			Cores:       int(cores),
		}, nil
	}

	var lpCores, ppCores uint32
	if perfLevels >= 1 {
		lpCores, err = darwinsys.SysctlUint32(lpKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get %v: %w", lpKey, err)
		}

		ppCores, err = darwinsys.SysctlUint32(ppKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get %v: %w", ppKey, err)
		}
	}

	var leCores, peCores uint32
	if perfLevels >= 2 {
		leCores, err = darwinsys.SysctlUint32(leKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get %v: %w", leKey, err)
		}

		peCores, err = darwinsys.SysctlUint32(peKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get %v: %w", peKey, err)
		}
	}

	return &CPU{
		BrandString:              brandString,
		Threads:                  int(threads),
		Cores:                    int(cores),
		LogicalPerformanceCores:  int(lpCores),
		LogicalEfficiencyCores:   int(leCores),
		PhysicalPerformanceCores: int(ppCores),
		PhysicalEfficiencyCores:  int(peCores),
	}, nil
}
