package dictionary

const Name = "dictionary"
const ManifestPath = "modules/dictionary/manifest.yaml"

type Module struct {
	Name         string
	ManifestPath string
}

func NewModule() Module {
	return Module{Name: Name, ManifestPath: ManifestPath}
}
