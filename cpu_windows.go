//go:build windows

package xpe

import (
	"fmt"
	"math/bits"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// LOGICAL_PROCESSOR_RELATIONSHIP enum values.
// https://learn.microsoft.com/en-us/windows/win32/api/winnt/ne-winnt-logical_processor_relationship
//
// We only need the relationProcessorCore value for current purposes, but define
// the entire enum in case we need it to extend functionality in the future.
const (
	relationProcessorCore    = iota // logical processors share a single processor core
	relationNumaNode                // logical processors are part of the same NUMA node
	relationCache                   // logical processors share a cache
	relationProcessorPackage        // logical processors share a physical package
	relationGroup                   // logical processors share a processor group
	relationProcessorDie            // logical processors share a single processor die
	relationNumaNodeEx              // RelationNumaNode with full group information
	relationProcessorModule         // logical processors share a processor module

	relationAll = 0xffff // retrieve all possible relationship types
)

// Header for SYSTEM_LOGICAL_PROCESSOR_INFORMATION_EX
// https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-system_logical_processor_information_ex
// we only need the beginning fields here.
type SYSTEM_LOGICAL_PROCESSOR_INFORMATION_EX struct {
	Relationship uint32
	Size         uint32
}

// Processor relationship struct (simplified)
// https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-processor_relationship
// actually contains more fields, but we only care about Flags and ProcessorMask.
type PROCESSOR_RELATIONSHIP struct {
	Flags           byte
	EfficiencyClass byte
	Reserved        [20]byte
	ProcessorMask   uint64
}

// getLogicalProcessorInfo calls the Windows API and returns the raw buffer.
func getLogicalProcessorInfo() ([]byte, error) {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	proc := kernel32.NewProc("GetLogicalProcessorInformationEx")

	var size uint32
	r1, _, lastErr := proc.Call(
		uintptr(relationProcessorCore),
		0,
		uintptr(unsafe.Pointer(&size)),
	)
	if r1 == 0 && lastErr != windows.ERROR_INSUFFICIENT_BUFFER {
		return nil, fmt.Errorf("initial syscall failed: %w", lastErr)
	}

	buf := make([]byte, size)
	r1, _, lastErr = proc.Call(
		uintptr(relationProcessorCore),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if r1 == 0 {
		return nil, fmt.Errorf("second syscall failed: %w", lastErr)
	}

	return buf, nil
}

// GetCPU returns information about the processor on Windows systems.
// It uses GetLogicalProcessorInformationEx to determine the number of
// performance and efficiency cores and threads, and reads the processor name
// from the registry if available.
func GetCPU() (*CPU, error) {
	buf, err := getLogicalProcessorInfo()
	if err != nil {
		return nil, err
	}

	type coreInfo struct {
		efficiencyClass byte
		threadCount     int
	}
	var cores []coreInfo
	var maxEffClass byte

	offset := 0
	for offset < len(buf) {
		header := (*SYSTEM_LOGICAL_PROCESSOR_INFORMATION_EX)(unsafe.Pointer(&buf[offset]))
		if header.Relationship == relationProcessorCore {
			pr := (*PROCESSOR_RELATIONSHIP)(unsafe.Add(unsafe.Pointer(&buf[offset]), unsafe.Sizeof(*header)))
			ci := coreInfo{
				efficiencyClass: pr.EfficiencyClass,
				threadCount:     bits.OnesCount64(pr.ProcessorMask),
			}
			if ci.efficiencyClass > maxEffClass {
				maxEffClass = ci.efficiencyClass
			}
			cores = append(cores, ci)
		}
		offset += int(header.Size)
	}

	// Classify cores. Cores with the highest EfficiencyClass are
	// performance cores; all others are efficiency cores. On homogeneous CPUs
	// (all cores same class), every core is treated as a performance core.
	var pCores, eCores, pThreads, eThreads int
	for _, ci := range cores {
		if ci.efficiencyClass == maxEffClass {
			pCores++
			pThreads += ci.threadCount
		} else {
			eCores++
			eThreads += ci.threadCount
		}
	}

	// try to fetch brand string from registry
	var brand string
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\CentralProcessor\0`, registry.QUERY_VALUE); err == nil {
		if s, _, err2 := k.GetStringValue("ProcessorNameString"); err2 == nil {
			brand = strings.TrimSpace(s)
		}
		k.Close()
	}

	return &CPU{
		BrandString:              brand,
		Threads:                  pThreads + eThreads,
		Cores:                    pCores + eCores,
		LogicalPerformanceCores:  pThreads,
		LogicalEfficiencyCores:   eThreads,
		PhysicalPerformanceCores: pCores,
		PhysicalEfficiencyCores:  eCores,
	}, nil
}
