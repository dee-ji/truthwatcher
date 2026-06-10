package discovery

import (
	"fmt"

	"truthwatcher/internal/policy"
)

const (
	ProfileJuniperJunos = "juniper_junos"
	ProfileCiscoIOSXR   = "cisco_iosxr"
)

// CommandMapping binds one vendor-specific read-only command to parser hints.
type CommandMapping struct {
	Command     string
	ParserHints []string
}

// Profile maps abstract discovery tasks to platform-specific read-only commands.
type Profile struct {
	Name     string
	Platform string
	Vendor   string
	Tasks    map[policy.Task][]CommandMapping
}

// CommandsForTask returns the commands a profile allows for an abstract task.
func (p Profile) CommandsForTask(task policy.Task) ([]CommandMapping, error) {
	commands, ok := p.Tasks[task]
	if !ok || len(commands) == 0 {
		return nil, fmt.Errorf("discovery profile %q has no commands for task %q", p.Name, task)
	}

	return cloneCommandMappings(commands), nil
}

// Validate checks that all profile tasks and commands are allowed by policy.
func (p Profile) Validate(engine policy.Engine) error {
	if p.Name == "" {
		return fmt.Errorf("discovery profile name is required")
	}
	if p.Platform == "" {
		return fmt.Errorf("discovery profile %q platform is required", p.Name)
	}
	if p.Vendor == "" {
		return fmt.Errorf("discovery profile %q vendor is required", p.Name)
	}
	if len(p.Tasks) == 0 {
		return fmt.Errorf("discovery profile %q must define tasks", p.Name)
	}

	for task, commands := range p.Tasks {
		if err := engine.CheckTask(task); err != nil {
			return fmt.Errorf("discovery profile %q task %q: %w", p.Name, task, err)
		}
		if len(commands) == 0 {
			return fmt.Errorf("discovery profile %q task %q must define commands", p.Name, task)
		}
		for _, command := range commands {
			if err := engine.CheckAction(policy.Action{Task: task, Command: command.Command}); err != nil {
				return fmt.Errorf("discovery profile %q task %q command %q: %w", p.Name, task, command.Command, err)
			}
		}
	}

	return nil
}

// BuiltInProfiles returns the compile-time discovery profiles shipped with the binary.
func BuiltInProfiles() map[string]Profile {
	profiles := map[string]Profile{
		ProfileJuniperJunos: juniperJunosProfile(),
		ProfileCiscoIOSXR:   ciscoIOSXRProfile(),
	}

	out := make(map[string]Profile, len(profiles))
	for name, profile := range profiles {
		out[name] = cloneProfile(profile)
	}
	return out
}

// BuiltInProfile returns one built-in profile by name.
func BuiltInProfile(name string) (Profile, bool) {
	profiles := BuiltInProfiles()
	profile, ok := profiles[name]
	return profile, ok
}

func juniperJunosProfile() Profile {
	return Profile{
		Name:     ProfileJuniperJunos,
		Platform: "junos",
		Vendor:   "juniper",
		Tasks: map[policy.Task][]CommandMapping{
			policy.TaskIdentifyDevice: {
				{Command: "show version", ParserHints: []string{"junos_version"}},
			},
			policy.TaskGetInventory: {
				{Command: "show chassis hardware", ParserHints: []string{"junos_chassis_hardware"}},
			},
			policy.TaskGetInterfaces: {
				{Command: "show interfaces terse", ParserHints: []string{"junos_interfaces_terse"}},
			},
			policy.TaskGetNeighbors: {
				{Command: "show lldp neighbors", ParserHints: []string{"junos_lldp_neighbors"}},
			},
			policy.TaskGetARP: {
				{Command: "show arp no-resolve", ParserHints: []string{"junos_arp"}},
			},
			policy.TaskGetIPv6Neighbors: {
				{Command: "show ipv6 neighbors", ParserHints: []string{"junos_ipv6_neighbors"}},
			},
			policy.TaskGetBGPSummary: {
				{Command: "show bgp summary", ParserHints: []string{"junos_bgp_summary"}},
			},
			policy.TaskGetRoutes: {
				{Command: "show route summary", ParserHints: []string{"junos_route_summary"}},
			},
		},
	}
}

func ciscoIOSXRProfile() Profile {
	return Profile{
		Name:     ProfileCiscoIOSXR,
		Platform: "iosxr",
		Vendor:   "cisco",
		Tasks: map[policy.Task][]CommandMapping{
			policy.TaskIdentifyDevice: {
				{Command: "show version", ParserHints: []string{"iosxr_version"}},
			},
			policy.TaskGetInventory: {
				{Command: "show inventory", ParserHints: []string{"iosxr_inventory"}},
			},
			policy.TaskGetInterfaces: {
				{Command: "show interfaces brief", ParserHints: []string{"iosxr_interfaces_brief"}},
			},
			policy.TaskGetNeighbors: {
				{Command: "show lldp neighbors", ParserHints: []string{"iosxr_lldp_neighbors"}},
				{Command: "show cdp neighbors", ParserHints: []string{"iosxr_cdp_neighbors"}},
			},
			policy.TaskGetARP: {
				{Command: "show arp", ParserHints: []string{"iosxr_arp"}},
			},
			policy.TaskGetIPv6Neighbors: {
				{Command: "show ipv6 neighbors", ParserHints: []string{"iosxr_ipv6_neighbors"}},
			},
			policy.TaskGetBGPSummary: {
				{Command: "show bgp summary", ParserHints: []string{"iosxr_bgp_summary"}},
			},
			policy.TaskGetRoutes: {
				{Command: "show route summary", ParserHints: []string{"iosxr_route_summary"}},
			},
		},
	}
}

func cloneProfile(profile Profile) Profile {
	clone := Profile{
		Name:     profile.Name,
		Platform: profile.Platform,
		Vendor:   profile.Vendor,
		Tasks:    make(map[policy.Task][]CommandMapping, len(profile.Tasks)),
	}

	for task, commands := range profile.Tasks {
		clone.Tasks[task] = cloneCommandMappings(commands)
	}

	return clone
}

func cloneCommandMappings(commands []CommandMapping) []CommandMapping {
	copied := make([]CommandMapping, len(commands))
	for i, command := range commands {
		copied[i] = CommandMapping{
			Command:     command.Command,
			ParserHints: append([]string(nil), command.ParserHints...),
		}
	}
	return copied
}
