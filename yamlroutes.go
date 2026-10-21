// Package yamlroutes — декларативные маршруты из YAML за реестром обработчиков.
//
// YAML знает только имена, код регистрирует реализации:
//
//	routes:
//	  - method: GET
//	    path: /api/v1/users
//	    do: UsersList
//
//	reg := yamlroutes.NewRegistry[gin.HandlerFunc]()
//	reg.RegisterHandler("UsersList", handlers.List)
//	err := reg.MountFile(router, "routes.yaml")
//
// Ядро пакета не зависит от фреймворка: обработчик — любой тип H,
// роутер — любой тип с Handle(method, path, H). Готовые адаптеры:
// ginroutes (gin) и nethttp (net/http ServeMux) в подпакетах.
//
// Несовпадение YAML и кода в любую сторону — ошибка при старте,
// а не 404 в проде.
package yamlroutes

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Route — одна ручка из YAML.
type Route struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
	Do     string `yaml:"do"`
}

// Spec — корневой документ YAML.
type Spec struct {
	Routes []Route `yaml:"routes"`
}

// Parse разбирает YAML и валидирует структуру (без привязки к реестру).
func Parse(data []byte) (Spec, error) {
	var spec Spec
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(true) // опечатка в ключе YAML — ошибка, а не тихий игнор
	if err := dec.Decode(&spec); err != nil {
		return Spec{}, fmt.Errorf("yamlroutes: parse: %w", err)
	}
	if err := spec.Validate(); err != nil {
		return Spec{}, err
	}
	return spec, nil
}

// LoadFile читает и парсит YAML-файл.
func LoadFile(path string) (Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, fmt.Errorf("yamlroutes: read %s: %w", path, err)
	}
	return Parse(data)
}

// Validate проверяет методы, пути и дубликаты.
func (s Spec) Validate() error {
	seen := make(map[string]bool, len(s.Routes))
	for i, r := range s.Routes {
		where := fmt.Sprintf("routes[%d] (%s %s -> %s)", i, r.Method, r.Path, r.Do)
		switch r.Method {
		case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
			http.MethodDelete, http.MethodHead, http.MethodOptions:
		default:
			return fmt.Errorf("yamlroutes: %s: unknown method %q", where, r.Method)
		}
		if !strings.HasPrefix(r.Path, "/") {
			return fmt.Errorf("yamlroutes: %s: path must start with /", where)
		}
		if strings.TrimSpace(r.Do) == "" {
			return fmt.Errorf("yamlroutes: %s: do is empty", where)
		}
		key := r.Method + " " + r.Path
		if seen[key] {
			return fmt.Errorf("yamlroutes: %s: duplicate route", where)
		}
		seen[key] = true
	}
	return nil
}

// Router — то, на что вешаются маршруты (адаптеры — в подпакетах).
type Router[H any] interface {
	Handle(method, path string, handler H)
}

// Registry — имя обработчика -> обработчик типа H.
type Registry[H any] struct {
	handlers map[string]H
}

func NewRegistry[H any]() *Registry[H] {
	return &Registry[H]{handlers: make(map[string]H)}
}

// RegisterHandler регистрирует один обработчик по имени из YAML.
func (r *Registry[H]) RegisterHandler(name string, h H) *Registry[H] {
	if _, dup := r.handlers[name]; dup {
		panic("yamlroutes: RegisterHandler(" + name + "): already registered")
	}
	r.handlers[name] = h
	return r
}

// RegisterHandlers — пакетная регистрация картой имя -> обработчик.
func (r *Registry[H]) RegisterHandlers(m map[string]H) *Registry[H] {
	for name, h := range m {
		r.RegisterHandler(name, h)
	}
	return r
}

// Get возвращает обработчик по имени.
func (r *Registry[H]) Get(name string) (H, bool) {
	h, ok := r.handlers[name]
	return h, ok
}

// Mount вешает все маршруты spec на router.
// Ошибка, если в YAML есть имя без зарегистрированного обработчика.
func (r *Registry[H]) Mount(router Router[H], spec Spec) error {
	if err := spec.Validate(); err != nil {
		return err
	}
	var missing []string
	for _, route := range spec.Routes {
		h, ok := r.handlers[route.Do]
		if !ok {
			missing = append(missing, route.Do)
			continue
		}
		router.Handle(route.Method, route.Path, h)
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("yamlroutes: handlers not registered: %s", strings.Join(missing, ", "))
	}
	return nil
}

// MountYAML — Parse + Mount одним вызовом.
func (r *Registry[H]) MountYAML(router Router[H], data []byte) error {
	spec, err := Parse(data)
	if err != nil {
		return err
	}
	return r.Mount(router, spec)
}

// MountFile — LoadFile + Mount одним вызовом.
func (r *Registry[H]) MountFile(router Router[H], path string) error {
	spec, err := LoadFile(path)
	if err != nil {
		return err
	}
	return r.Mount(router, spec)
}
