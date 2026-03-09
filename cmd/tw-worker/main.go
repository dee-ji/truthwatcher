package main

import (
	"fmt"

	"github.com/truthwatcher/truthwatcher/pkg/version"
)

func main() {
	// TODO: Remove legacy wrapper after external scripts migrate to cmd/squire.
	fmt.Printf("tw-worker is deprecated; run squire instead (version %s)\n", version.Version)
}
