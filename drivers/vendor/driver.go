package vendor

type Renderer interface {
	Name() string
	Render(target string, ir map[string]string) (string, error)
}
