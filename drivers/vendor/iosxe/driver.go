package iosxe

import "fmt"

type Driver struct{}

func (Driver) Name() string { return "iosxe" }
func (Driver) Render(target string, ir map[string]string) (string, error) {
	return fmt.Sprintf("! vendor=iosxe target=%s\nhostname %s\n", target, ir["hostname"]), nil
}
