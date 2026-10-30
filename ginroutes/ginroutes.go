// Package ginroutes — адаптер yamlroutes для gin.
package ginroutes

import (
	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/yamlroutes"
)

// Registry — реестр gin-обработчиков.
type Registry = yamlroutes.Registry[gin.HandlerFunc]

// NewRegistry создаёт пустой реестр.
func NewRegistry() *Registry { return yamlroutes.NewRegistry[gin.HandlerFunc]() }

type adapter struct{ r gin.IRouter }

func (a adapter) Handle(route yamlroutes.Route, h gin.HandlerFunc) {
	a.r.Handle(route.Method, route.Path, h)
}

// Mount вешает spec на gin-роутер.
func Mount(r gin.IRouter, reg *Registry, spec yamlroutes.Spec) error {
	return reg.Mount(adapter{r}, spec)
}

// MountYAML парсит YAML и вешает маршруты на gin-роутер.
func MountYAML(r gin.IRouter, reg *Registry, data []byte) error {
	return reg.MountYAML(adapter{r}, data)
}

// MountFile читает YAML-файл и вешает маршруты на gin-роутер.
func MountFile(r gin.IRouter, reg *Registry, path string) error {
	return reg.MountFile(adapter{r}, path)
}
