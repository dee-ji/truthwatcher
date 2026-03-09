package eos

import "fmt"

type Driver struct{}

func (Driver) Name() string { return "eos" }
func (Driver) Render(target string, ir map[string]string) (string, error) {
	return fmt.Sprintf("! vendor=eos target=%s\nhostname %s\n", target, ir["hostname"]), nil
}
