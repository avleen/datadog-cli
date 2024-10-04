package modules

var modules = make(map[string]Module)

func RegisterModule(m Module) {
	modules[m.Name()] = m
}

func GetModules() []string {
	var moduleNames []string
	for name := range modules {
		moduleNames = append(moduleNames, name)
	}
	return moduleNames
}

func GetModule(name string) (Module, bool) {
	m, exists := modules[name]
	return m, exists
}

func RegisterAllModules() {
	RegisterModule(NewExampleModule())
	RegisterModule(NewMetricsModule())
	RegisterModule(NewContainersModule())
	RegisterModule(NewHostsModule())
}
