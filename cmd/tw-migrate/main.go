package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: tw-migrate [up|down|status]")
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "up", "down", "status":
		fmt.Printf("tw-migrate %s: scaffold command executed\n", cmd)
		fmt.Println("TODO(truthwatcher): connect to migration engine and apply SQL files from ./migrations")
	default:
		fmt.Printf("unknown command %q\n", cmd)
		fmt.Println("usage: tw-migrate [up|down|status]")
	}
}
