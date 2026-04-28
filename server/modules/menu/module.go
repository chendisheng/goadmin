package menu

const Name = "menu"
const ManifestPath = "modules/menu/manifest.yaml"

type Module struct {
	Name         string
	ManifestPath string
}

func NewModule() Module { return Module{Name: Name, ManifestPath: ManifestPath} }
