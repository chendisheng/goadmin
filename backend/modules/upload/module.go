package upload

const Name = "upload"
const ManifestPath = "modules/upload/manifest.yaml"

type Module struct {
	Name         string
	ManifestPath string
}

func NewModule() Module {
	return Module{Name: Name, ManifestPath: ManifestPath}
}
