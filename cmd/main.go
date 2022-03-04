package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/mroth/xpe"
)

func main() {
	cpu, err := xpe.GetCPU()
	if err != nil {
		log.Fatal(err)
	}

	s, _ := json.MarshalIndent(cpu, "", "  ")
	fmt.Printf("%s\n", s)
}
