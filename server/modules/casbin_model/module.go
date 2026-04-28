package casbin_model

const Name = "casbin_model"
const ManifestPath = "modules/casbin_model/manifest.yaml"

type Module struct {
	Name         string
	ManifestPath string
}

func NewModule() Module {
	return Module{Name: Name, ManifestPath: ManifestPath}
}
