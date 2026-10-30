package ginroutes

import (
	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/ginserver"
	"github.com/SaveliiYam/yamlroutes"
)

// HandlerRegistry — реестр ЧИСТЫХ обработчиков ginserver.HandlerFunc.
type HandlerRegistry struct {
	reg *yamlroutes.Registry[ginserver.HandlerFunc]
}

func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{reg: yamlroutes.NewRegistry[ginserver.HandlerFunc]()}
}

func (r *HandlerRegistry) RegisterHandler(name string, h ginserver.HandlerFunc) *HandlerRegistry {
	r.reg.RegisterHandler(name, h)
	return r
}

func (r *HandlerRegistry) RegisterHandlers(m map[string]ginserver.HandlerFunc) *HandlerRegistry {
	r.reg.RegisterHandlers(m)
	return r
}

type pureAdapter struct {
	r             gin.IRouter
	basicAccounts ginserver.Accounts
	basicOpts     []ginserver.BasicAuthOption
}

func (a pureAdapter) Handle(route yamlroutes.Route, h ginserver.HandlerFunc) {
	wrapped := ginserver.Wrap(h)
	if route.HasAuth("basic") {
		// Per-route: без глобальных skipPaths — basic обязателен на этой ручке.
		opts := append([]ginserver.BasicAuthOption{ginserver.WithSkipPaths()}, a.basicOpts...)
		a.r.Handle(route.Method, route.Path,
			ginserver.BasicAuth(a.basicAccounts, opts...),
			wrapped,
		)
		return
	}
	a.r.Handle(route.Method, route.Path, wrapped)
}

// Mount вешает spec; basicAccounts нужны, если в YAML есть auth: [basic].
func (r *HandlerRegistry) Mount(router gin.IRouter, spec yamlroutes.Spec, basicAccounts ginserver.Accounts, basicOpts ...ginserver.BasicAuthOption) error {
	if spec.RequiresAuth("basic") && len(basicAccounts) == 0 {
		panic("ginroutes: routes declare auth: basic but basic_auth accounts are empty")
	}
	return r.reg.Mount(pureAdapter{r: router, basicAccounts: basicAccounts, basicOpts: basicOpts}, spec)
}

// MountYAML парсит YAML и вешает маршруты.
func (r *HandlerRegistry) MountYAML(router gin.IRouter, data []byte, basicAccounts ginserver.Accounts, basicOpts ...ginserver.BasicAuthOption) error {
	spec, err := yamlroutes.Parse(data)
	if err != nil {
		return err
	}
	return r.Mount(router, spec, basicAccounts, basicOpts...)
}

// MountFile читает YAML-файл и вешает маршруты.
func (r *HandlerRegistry) MountFile(router gin.IRouter, path string, basicAccounts ginserver.Accounts, basicOpts ...ginserver.BasicAuthOption) error {
	spec, err := yamlroutes.LoadFile(path)
	if err != nil {
		return err
	}
	return r.Mount(router, spec, basicAccounts, basicOpts...)
}

// MountHandlersFile — зарегистрировать map обработчиков и смонтировать routes.yaml.
//
// Если в YAML у ручки указано auth: [basic], а basicAccounts пуст — panic при старте.
// Если auth не указан — basic для ручки не применяется.
func MountHandlersFile(router gin.IRouter, path string, handlers map[string]ginserver.HandlerFunc, basicAccounts ginserver.Accounts, basicOpts ...ginserver.BasicAuthOption) error {
	reg := NewHandlerRegistry()
	reg.RegisterHandlers(handlers)
	return reg.MountFile(router, path, basicAccounts, basicOpts...)
}
