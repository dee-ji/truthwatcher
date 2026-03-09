package iosxr

import "fmt"

type Driver struct{}

func (Driver) Name() string { return "iosxr" }
func (Driver) Render(target string, ir map[string]string) (string, error) {
	return fmt.Sprintf("! vendor=iosxr target=%s\nhostname %s\n", target, ir["hostname"]), nil
}
