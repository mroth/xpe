package xpe

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type mockSysctler struct {
	data map[string]string
}

func newMockSysctler(t *testing.T, path string) mockSysctler {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	data := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if i := strings.IndexRune(line, '='); i > 0 {
			k, v := line[:i], line[i+1:]
			data[k] = v
		} else {
			t.Logf("malformed testdata line: %s", line)
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	return mockSysctler{data: data}
}

func (s mockSysctler) Sysctl(name string) (value string, err error) {
	value, ok := s.data[name]
	if !ok {
		err = fmt.Errorf("not found: %v", name)
	}
	return
}

func (s mockSysctler) SysctlUint32(name string) (uint32, error) {
	value, ok := s.data[name]
	if !ok {
		return 0, fmt.Errorf("not found: %v", name)
	}
	i, err := strconv.Atoi(value)
	return uint32(i), err
}

func TestMockGetCPU_Darwin(t *testing.T) {
	var mocks = []struct {
		name string
		path string
		cpu  *CPU
	}{
		{
			name: "MacBook Pro (14-inch, 2021) - Apple M1 Pro [8p, 2e, 16g], macOS 12.2.1",
			path: "testdata/m1pro-macos12_2_1.txt",
			cpu: &CPU{
				BrandString:              "Apple M1 Pro",
				Threads:                  10,
				Cores:                    10,
				LogicalPerformanceCores:  8,
				LogicalEfficiencyCores:   2,
				PhysicalPerformanceCores: 8,
				PhysicalEfficiencyCores:  2,
			},
		},
		{
			name: "MacBook Pro (13-inch, 2019, Four Thunderbolt 3 ports) - 2.8 GHz Quad-Core Intel Core i7 [4p HT], macOS 12.2.1",
			path: "testdata/mbpro-2019i7-macos12_2_1.txt",
			cpu: &CPU{
				BrandString:              "Intel(R) Core(TM) i7-8569U CPU @ 2.80GHz",
				Threads:                  8,
				Cores:                    4,
				LogicalPerformanceCores:  8,
				LogicalEfficiencyCores:   0,
				PhysicalPerformanceCores: 4,
				PhysicalEfficiencyCores:  0,
			},
		},
		{
			name: "MacBook Air (M1, 2020) - Apple M1 [4p, 4e, 8g], macOS 12.2.1",
			path: "testdata/m1air2020-macos12_2_1.txt",
			cpu: &CPU{
				BrandString:              "Apple M1",
				Threads:                  8,
				Cores:                    8,
				LogicalPerformanceCores:  4,
				LogicalEfficiencyCores:   4,
				PhysicalPerformanceCores: 4,
				PhysicalEfficiencyCores:  4,
			},
		},
	}

	defer func() { darwinsys = nativeSyscallPackage{} }() // return to normal
	for _, m := range mocks {
		t.Run(m.path, func(t *testing.T) {
			darwinsys = newMockSysctler(t, m.path)
			cpu, err := GetCPU()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(m.cpu, cpu) {
				t.Fatalf("want %v, got %v", m.cpu, cpu)
			}
		})
	}
}
