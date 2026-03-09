package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/truthwatcher/truthwatcher/pkg/version"
)

func main() {
	cmd := strings.Join(os.Args[1:], " ")
	switch cmd {
	case "intent validate", "intent diff", "topology query", "deploy create", "deploy get", "state compare", "render preview":
		fmt.Printf("%s: scaffold\n", cmd)
	case "version":
		fmt.Println(version.Version)
	default:
		fmt.Println("usage: twctl [intent validate|intent diff|topology query|deploy create|deploy get|state compare|render preview|version]")
	}
}
