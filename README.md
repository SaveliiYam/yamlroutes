# yamlroutes — маршруты из YAML

Декларативные HTTP-маршруты: YAML знает `method + path + do: ИмяОбработчика`,
код регистрирует реализации по именам. Несовпадение в любую сторону —
ошибка при старте, а не 404 в проде.

Ядро не зависит от фреймворка (обработчик — любой тип, роутер — любой
`Handle(method, path, H)`). Адаптеры: `ginroutes` (gin) и `nethttp`
(стандартный `http.ServeMux`, конвертирует `/users/:id` в `/users/{id}`).

## Установка

```bash
go get github.com/SaveliiYam/yamlroutes@v0.1.0
```

## Использование (gin)

`routes.yaml`:

```yaml
routes:
  - method: GET
    path: /api/v1/users
    do: UsersList
  - method: POST
    path: /api/v1/users
    do: UserCreate
```

Код:

```go
reg := ginroutes.NewRegistry()
reg.RegisterHandler("UsersList", handlers.List)
// или пакетом:
reg.RegisterHandlers(map[string]gin.HandlerFunc{
	"UserCreate": handlers.Create,
})

err := ginroutes.MountFile(engine, reg, "routes.yaml") // или MountYAML(engine, reg, data)
```

Валидация при старте: неизвестный метод, путь без `/`, пустой `do`,
опечатка в ключе YAML, дубликат `method+path`, незарегистрированное имя.

Требуется Go ≥ 1.25.
