// Package xpe contains experimental code for determining the number of
// performance and efficiency cores on the runtime CPU architecture.
package xpe

type CPU struct {
	BrandString              string
	Threads                  int
	Cores                    int
	LogicalPerformanceCores  int
	LogicalEfficiencyCores   int
	PhysicalPerformanceCores int
	PhysicalEfficiencyCores  int
}
