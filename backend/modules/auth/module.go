package auth

const Name = "auth"

const ManifestPath = "modules/auth/manifest.yaml"

type Module struct {
	Name         string
	ManifestPath string
}

func NewModule() Module {
	return Module{
		Name:         Name,
		ManifestPath: ManifestPath,
	}
}
