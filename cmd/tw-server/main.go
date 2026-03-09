package main

import (
	"fmt"

	"github.com/truthwatcher/truthwatcher/pkg/version"
)

func main() {
	// TODO: Remove legacy wrapper after external scripts migrate to cmd/spanreed.
	fmt.Printf("tw-server is deprecated; run spanreed instead (version %s)\n", version.Version)
}
