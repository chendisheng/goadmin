package bootstrap

func Modules() []Module {
	modules := builtinModules()
	return append(modules, generatedModules()...)
}

func GeneratedModules() []Module {
	return generatedModules()
}
