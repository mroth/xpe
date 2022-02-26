package xpe

import (
	"fmt"
	"syscall"
)

//nperflevels
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

	brandString, err := syscall.Sysctl(cpuBrandString)
	if err != nil {
		return nil, fmt.Errorf("failed to get %v: %w", cpuBrandString, err)
	}

	threads, err := syscall.SysctlUint32(cpuThreadCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get %v: %w", cpuThreadCount, err)
	}

	cores, err := syscall.SysctlUint32(cpuCoreCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get %v: %w", cpuCoreCount, err)
	}

	perfLevels, err := syscall.SysctlUint32(hwNPerflevels)
	if err != nil {
		// TODO: seems might not exist on older versions of macOS? lets not error if so
		return &CPU{
			BrandString: brandString,
			Threads:     int(threads),
			Cores:       int(cores),
		}, nil
	}

	var lpCores, leCores uint32
	if perfLevels >= 1 {
		lpCores, err = syscall.SysctlUint32(lpKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get %v: %w", lpKey, err)
		}

		leCores, err = syscall.SysctlUint32(leKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get %v: %w", leKey, err)
		}
	}

	var ppCores, peCores uint32
	if perfLevels >= 2 {
		ppCores, err = syscall.SysctlUint32(ppKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get %v: %w", ppKey, err)
		}

		peCores, err = syscall.SysctlUint32(peKey)
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
