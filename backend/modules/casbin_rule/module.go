package casbin_rule

const Name = "casbin_rule"
const ManifestPath = "modules/casbin_rule/manifest.yaml"

type Module struct {
	Name         string
	ManifestPath string
}

func NewModule() Module {
	return Module{Name: Name, ManifestPath: ManifestPath}
}
