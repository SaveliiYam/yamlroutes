# yamlroutes — маршруты из YAML

Декларативные HTTP-маршруты: YAML знает `method + path + do: ИмяОбработчика`
и опционально `auth`, код регистрирует реализации по именам. Несовпадение —
ошибка при старте, а не 404 в проде.

Ядро не зависит от фреймворка. Адаптеры: `ginroutes` (gin) и `nethttp`
(стандартный `http.ServeMux`, конвертирует `/users/:id` в `/users/{id}`).

## Установка

```bash
go get github.com/SaveliiYam/yamlroutes@v0.5.0
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
    auth: [basic]   # опционально; пока поддерживается basic
```

Код:

```go
reg := ginroutes.NewRegistry()
reg.RegisterHandler("UsersList", handlers.List)
reg.RegisterHandlers(map[string]gin.HandlerFunc{
	"UserCreate": handlers.Create,
})

err := ginroutes.MountFile(engine, reg, "routes.yaml")
```

Валидация при старте: неизвестный метод, путь без `/`, пустой `do`,
неизвестный `auth`, опечатка в ключе YAML, дубликат `method+path`,
незарегистрированное имя.

Требуется Go ≥ 1.25.

## Auth на ручке

Поле `auth` — список способов авторизации для конкретной ручки.
Сейчас поддерживается только `basic`.

- `auth` не указан / пуст — basic для ручки не применяется.
- `auth: [basic]` — на ручку вешается Basic Auth; при монтировании через
  `HandlerRegistry` / `MountHandlersFile` нужны непустые accounts,
  иначе **panic** при старте (конфиг basic обязателен).

```go
err := ginroutes.MountHandlersFile(engine, "routes.yaml", handlers, accounts,
	ginserver.WithRealm("jwt"),
)
```

## Чистые обработчики для gin (с v0.2.0)

`ginroutes.HandlerRegistry` принимает `ginserver.HandlerFunc` и сам
оборачивает их в конверт при монтировании:

```go
reg := ginroutes.NewHandlerRegistry()
reg.RegisterHandler("UsersList", handlers.List)
err := reg.MountFile(engine, "routes.yaml", accounts)

// или одной строкой:
err := ginroutes.MountHandlersFile(engine, "routes.yaml", map[string]ginserver.HandlerFunc{
	"UsersList": handlers.List,
}, accounts)
```
