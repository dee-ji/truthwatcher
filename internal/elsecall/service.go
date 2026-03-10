package elsecall

import "fmt"

type Artifact struct {
	Vendor   string
	Format   string
	Rendered string
}

type IntermediateArtifact struct {
	Hostname string
	Role     string
}

type CompilerService struct{}

func NewCompilerService() *CompilerService { return &CompilerService{} }

func (s *CompilerService) BuildIntermediate(spec map[string]any) IntermediateArtifact {
	hostname := "unnamed-device"
	if metadata, ok := spec["metadata"].(map[string]any); ok {
		if name, ok := metadata["name"].(string); ok && name != "" {
			hostname = name
		}
	}
	return IntermediateArtifact{Hostname: hostname, Role: "leaf"}
}

func (s *CompilerService) RenderJunos(ir IntermediateArtifact) Artifact {
	return Artifact{
		Vendor:   "junos",
		Format:   "set",
		Rendered: fmt.Sprintf("set system host-name %s\nset routing-options autonomous-system 65000\n", ir.Hostname),
	}
}
