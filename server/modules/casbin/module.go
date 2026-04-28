package casbin

const Name = "casbin"
const ManifestPath = "modules/casbin/manifest.yaml"

type Module struct {
	Name         string
	ManifestPath string
}

func NewModule() Module {
	return Module{Name: Name, ManifestPath: ManifestPath}
}
