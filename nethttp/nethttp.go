// Package nethttp — адаптер yamlroutes для стандартного http.ServeMux.
//
// Пути в YAML пишутся в gin-стиле (/users/:id, /files/*path) и
// конвертируются в паттерны ServeMux (/users/{id}, /files/{path...}).
package nethttp

import (
	"net/http"
	"strings"

	"github.com/SaveliiYam/yamlroutes"
)

// Registry — реестр net/http-обработчиков.
type Registry = yamlroutes.Registry[http.HandlerFunc]

// NewRegistry создаёт пустой реестр.
func NewRegistry() *Registry { return yamlroutes.NewRegistry[http.HandlerFunc]() }

type adapter struct{ mux *http.ServeMux }

func (a adapter) Handle(route yamlroutes.Route, h http.HandlerFunc) {
	a.mux.HandleFunc(route.Method+" "+ConvertPath(route.Path), h)
}

// Mount вешает spec на ServeMux.
func Mount(mux *http.ServeMux, reg *Registry, spec yamlroutes.Spec) error {
	return reg.Mount(adapter{mux}, spec)
}

// MountYAML парсит YAML и вешает маршруты на ServeMux.
func MountYAML(mux *http.ServeMux, reg *Registry, data []byte) error {
	return reg.MountYAML(adapter{mux}, data)
}

// MountFile читает YAML-файл и вешает маршруты на ServeMux.
func MountFile(mux *http.ServeMux, reg *Registry, path string) error {
	return reg.MountFile(adapter{mux}, path)
}

// ConvertPath: /users/:id -> /users/{id}, /files/*path -> /files/{path...}.
func ConvertPath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		switch {
		case strings.HasPrefix(p, ":"):
			parts[i] = "{" + strings.TrimPrefix(p, ":") + "}"
		case strings.HasPrefix(p, "*"):
			parts[i] = "{" + strings.TrimPrefix(p, "*") + "...}"
		}
	}
	return strings.Join(parts, "/")
}
