package junos

import "fmt"

type Driver struct{}

func (Driver) Name() string { return "junos" }
func (Driver) Render(target string, ir map[string]string) (string, error) {
	return fmt.Sprintf("! vendor=junos target=%s\nhostname %s\n", target, ir["hostname"]), nil
}
