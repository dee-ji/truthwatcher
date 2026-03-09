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
	fmt.Printf("migration command %s executed (scaffold)\n", os.Args[1])
}
