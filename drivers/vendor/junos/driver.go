package junos

import "fmt"

type Driver struct{}

func (Driver) Name() string { return "junos" }
func (Driver) Render(target string, ir map[string]string) (string, error) {
	hostname := ir["hostname"]
	if hostname == "" {
		hostname = target
	}
	return fmt.Sprintf("set system host-name %s\nset routing-options autonomous-system 65000\n", hostname), nil
}
