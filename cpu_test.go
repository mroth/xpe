package xpe_test

import (
	"testing"

	"github.com/mroth/xpe"
)

func TestGetCPU(t *testing.T) {
	c, err := xpe.GetCPU()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", *c)
}

func BenchmarkGetCPU(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = xpe.GetCPU()
	}
}
