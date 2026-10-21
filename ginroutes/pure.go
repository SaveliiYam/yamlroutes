package ginroutes

import (
	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/ginserver"
	"github.com/SaveliiYam/yamlroutes"
)

// HandlerRegistry — реестр ЧИСТЫХ обработчиков ginserver.HandlerFunc
// (возвращают ginserver.Result через утилиты Data/Unprocessable/...).
// Wrap делается внутри при монтировании: в прикладном коде обёрток нет,
// статус каждый обработчик задаёт сам в возвращаемом Result.
//
//	reg := ginroutes.NewHandlerRegistry()
//	reg.RegisterHandler("UsersList", handlers.List)
//	reg.RegisterHandlers(map[string]ginserver.HandlerFunc{"UserGet": handlers.Get})
//	err := reg.MountFile(engine, "routes.yaml")
type HandlerRegistry struct {
	reg *yamlroutes.Registry[ginserver.HandlerFunc]
}

func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{reg: yamlroutes.NewRegistry[ginserver.HandlerFunc]()}
}

// RegisterHandler регистрирует чистый обработчик.
func (r *HandlerRegistry) RegisterHandler(name string, h ginserver.HandlerFunc) *HandlerRegistry {
	r.reg.RegisterHandler(name, h)
	return r
}

// RegisterHandlers — пакетная регистрация.
func (r *HandlerRegistry) RegisterHandlers(m map[string]ginserver.HandlerFunc) *HandlerRegistry {
	r.reg.RegisterHandlers(m)
	return r
}

type pureAdapter struct{ r gin.IRouter }

func (a pureAdapter) Handle(method, path string, h ginserver.HandlerFunc) {
	a.r.Handle(method, path, ginserver.Wrap(h))
}

// Mount вешает spec на gin-роутер, оборачивая чистые обработчики.
func (r *HandlerRegistry) Mount(router gin.IRouter, spec yamlroutes.Spec) error {
	return r.reg.Mount(pureAdapter{router}, spec)
}

// MountYAML парсит YAML и вешает маршруты.
func (r *HandlerRegistry) MountYAML(router gin.IRouter, data []byte) error {
	return r.reg.MountYAML(pureAdapter{router}, data)
}

// MountFile читает YAML-файл и вешает маршруты.
func (r *HandlerRegistry) MountFile(router gin.IRouter, path string) error {
	return r.reg.MountFile(pureAdapter{router}, path)
}
