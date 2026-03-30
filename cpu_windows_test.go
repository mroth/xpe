//go:build windows

package xpe

import (
	"reflect"
	"testing"
	"unsafe"
)

type mockWindowsProvider struct {
	buf   []byte
	brand string
}

func (m *mockWindowsProvider) getLogicalProcessorInfoBuf() ([]byte, error) {
	return m.buf, nil
}

func (m *mockWindowsProvider) getProcessorBrandString() string {
	return m.brand
}

type testCore struct {
	efficiencyClass byte
	threadMask      uint64
}

// buildProcessorInfoBuf constructs a raw byte buffer matching the layout that
// GetLogicalProcessorInformationEx returns, from a slice of core descriptors.
func buildProcessorInfoBuf(t *testing.T, cores []testCore) []byte {
	t.Helper()

	headerSize := unsafe.Sizeof(SYSTEM_LOGICAL_PROCESSOR_INFORMATION_EX{})
	prSize := unsafe.Sizeof(PROCESSOR_RELATIONSHIP{})
	entrySize := uint32(headerSize + prSize)

	var buf []byte
	for _, c := range cores {
		entry := make([]byte, entrySize)
		header := (*SYSTEM_LOGICAL_PROCESSOR_INFORMATION_EX)(unsafe.Pointer(&entry[0]))
		header.Relationship = relationProcessorCore
		header.Size = entrySize

		pr := (*PROCESSOR_RELATIONSHIP)(unsafe.Pointer(&entry[headerSize]))
		pr.EfficiencyClass = c.efficiencyClass
		pr.GroupCount = 1
		pr.GroupMask[0] = GROUP_AFFINITY{Mask: c.threadMask}

		buf = append(buf, entry...)
	}
	return buf
}

// repeat creates n copies of the given testCore.
func repeat(n int, c testCore) []testCore {
	cores := make([]testCore, n)
	for i := range cores {
		cores[i] = c
	}
	return cores
}

func TestGetCPU_Windows(t *testing.T) {
	tests := []struct {
		name  string
		cores []testCore
		brand string
		want  *CPU
	}{
		{
			name:  "Intel 13th Gen hybrid (8P HT + 8E)",
			cores: append(repeat(8, testCore{1, 0x3}), repeat(8, testCore{0, 0x1})...),
			brand: "13th Gen Intel(R) Core(TM) i7-13700K",
			want: &CPU{
				BrandString:              "13th Gen Intel(R) Core(TM) i7-13700K",
				Threads:                  24,
				Cores:                    16,
				LogicalPerformanceCores:  16,
				LogicalEfficiencyCores:   8,
				PhysicalPerformanceCores: 8,
				PhysicalEfficiencyCores:  8,
			},
		},
		{
			name:  "Homogeneous Intel (8 cores HT, all same efficiency class)",
			cores: repeat(8, testCore{0, 0x3}),
			brand: "Intel(R) Core(TM) i7-10700K CPU @ 3.80GHz",
			want: &CPU{
				BrandString:              "Intel(R) Core(TM) i7-10700K CPU @ 3.80GHz",
				Threads:                  16,
				Cores:                    8,
				LogicalPerformanceCores:  16,
				LogicalEfficiencyCores:   0,
				PhysicalPerformanceCores: 8,
				PhysicalEfficiencyCores:  0,
			},
		},
		{
			name:  "Snapdragon X Elite (8P + 4E, no HT)",
			cores: append(repeat(8, testCore{1, 0x1}), repeat(4, testCore{0, 0x1})...),
			brand: "Snapdragon(R) X Elite - X1E78100 - Qualcomm(R) Oryon(TM) CPU",
			want: &CPU{
				BrandString:              "Snapdragon(R) X Elite - X1E78100 - Qualcomm(R) Oryon(TM) CPU",
				Threads:                  12,
				Cores:                    12,
				LogicalPerformanceCores:  8,
				LogicalEfficiencyCores:   4,
				PhysicalPerformanceCores: 8,
				PhysicalEfficiencyCores:  4,
			},
		},
		{
			name:  "Single core CPU",
			cores: []testCore{{0, 0x1}},
			brand: "Some Single Core CPU",
			want: &CPU{
				BrandString:              "Some Single Core CPU",
				Threads:                  1,
				Cores:                    1,
				LogicalPerformanceCores:  1,
				LogicalEfficiencyCores:   0,
				PhysicalPerformanceCores: 1,
				PhysicalEfficiencyCores:  0,
			},
		},
	}

	defer func() { winsys = nativeWindowsProvider{} }()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			winsys = &mockWindowsProvider{
				buf:   buildProcessorInfoBuf(t, tt.cores),
				brand: tt.brand,
			}
			got, err := GetCPU()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Fatalf("want %+v, got %+v", tt.want, got)
			}
		})
	}
}