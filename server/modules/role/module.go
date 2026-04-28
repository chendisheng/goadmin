package role

const Name = "role"
const ManifestPath = "modules/role/manifest.yaml"

type Module struct {
	Name         string
	ManifestPath string
}

func NewModule() Module { return Module{Name: Name, ManifestPath: ManifestPath} }
