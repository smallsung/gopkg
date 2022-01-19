package rpc

const builtinServiceName = "rpc"

type builtinService struct {
	server *Server
}

func (s builtinService) Modules() []string {
	var modules []string
	for namespace, _ := range s.server.registry.services {
		modules = append(modules, namespace)
	}
	return modules
}

func (s builtinService) Module(namespace string) []string {
	methods := make([]string, 0)
	if service, exist := s.server.registry.services[namespace]; exist {
		for name, _ := range service.callbacks {
			methods = append(methods, name)
		}
	}
	return methods
}
