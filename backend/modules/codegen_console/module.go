package codegen_console

const Name = "codegen_console"
const ManifestPath = "modules/codegen_console/manifest.yaml"

type Module struct {
	Name         string
	ManifestPath string
}

func NewModule() Module {
	return Module{Name: Name, ManifestPath: ManifestPath}
}
