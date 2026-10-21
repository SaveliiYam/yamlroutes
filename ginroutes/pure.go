package ginroutes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/ginserver"
	"github.com/SaveliiYam/yamlroutes"
)

// HandlerRegistry — реестр ЧИСТЫХ обработчиков ginserver.HandlerFunc.
// Wrap делается внутри при монтировании: в прикладном коде обёрток нет.
//
//	reg := ginroutes.NewHandlerRegistry()
//	reg.RegisterHandler("UsersList", handlers.List)
//	reg.RegisterHandlers(map[string]ginserver.HandlerFunc{"UserGet": handlers.Get})
//	reg.SetStatus("UserCreate", http.StatusCreated)
//	err := reg.MountFile(engine, "routes.yaml")
type HandlerRegistry struct {
	reg      *yamlroutes.Registry[ginserver.HandlerFunc]
	statuses map[string]int
}

func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		reg:      yamlroutes.NewRegistry[ginserver.HandlerFunc](),
		statuses: make(map[string]int),
	}
}

// RegisterHandler регистрирует чистый обработчик (статус успеха 200).
func (r *HandlerRegistry) RegisterHandler(name string, h ginserver.HandlerFunc) *HandlerRegistry {
	r.reg.RegisterHandler(name, h)
	return r
}

// RegisterHandlerStatus — то же, но со статусом успеха (например, 201).
func (r *HandlerRegistry) RegisterHandlerStatus(name string, status int, h ginserver.HandlerFunc) *HandlerRegistry {
	r.reg.RegisterHandler(name, h)
	r.statuses[name] = status
	return r
}

// RegisterHandlers — пакетная регистрация (статус 200 у всех).
func (r *HandlerRegistry) RegisterHandlers(m map[string]ginserver.HandlerFunc) *HandlerRegistry {
	r.reg.RegisterHandlers(m)
	return r
}

// SetStatus задаёт статус успеха уже зарегистрированному обработчику.
func (r *HandlerRegistry) SetStatus(name string, status int) *HandlerRegistry {
	if _, ok := r.reg.Get(name); !ok {
		panic("ginroutes: SetStatus(" + name + "): handler not registered")
	}
	r.statuses[name] = status
	return r
}

type pureAdapter struct {
	r       gin.IRouter
	byRoute map[string]int // "METHOD path" -> status, заполняется в Mount*
}

// Handle вызывается из yamlroutes.Mount: имя обработчика здесь уже потеряно,
// поэтому статус подбирается по паре method+path из byRoute.
func (a pureAdapter) Handle(method, path string, h ginserver.HandlerFunc) {
	status := a.byRoute[method+" "+path]
	if status == 0 {
		status = http.StatusOK
	}
	a.r.Handle(method, path, ginserver.WrapStatus(status, h))
}

func (r *HandlerRegistry) adapter(router gin.IRouter, spec yamlroutes.Spec) pureAdapter {
	byRoute := make(map[string]int, len(spec.Routes))
	for _, route := range spec.Routes {
		if s, ok := r.statuses[route.Do]; ok {
			byRoute[route.Method+" "+route.Path] = s
		}
	}
	return pureAdapter{r: router, byRoute: byRoute}
}

// Mount вешает spec на gin-роутер, оборачивая чистые обработчики.
func (r *HandlerRegistry) Mount(router gin.IRouter, spec yamlroutes.Spec) error {
	return r.reg.Mount(r.adapter(router, spec), spec)
}

// MountYAML парсит YAML и вешает маршруты.
func (r *HandlerRegistry) MountYAML(router gin.IRouter, data []byte) error {
	spec, err := yamlroutes.Parse(data)
	if err != nil {
		return err
	}
	return r.Mount(router, spec)
}

// MountFile читает YAML-файл и вешает маршруты.
func (r *HandlerRegistry) MountFile(router gin.IRouter, path string) error {
	spec, err := yamlroutes.LoadFile(path)
	if err != nil {
		return err
	}
	return r.Mount(router, spec)
}
