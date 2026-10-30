package ginroutes_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/ginserver"
	"github.com/SaveliiYam/yamlroutes/ginroutes"
)

func TestHandlerRegistryWrapsPureHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reg := ginroutes.NewHandlerRegistry()
	// Чистые функции, возвращающие Result: ни одного Wrap в прикладном коде.
	reg.RegisterHandler("UsersList", func(*gin.Context) ginserver.Result {
		return ginserver.Data([]string{"a"})
	})
	reg.RegisterHandler("UserCreate", func(*gin.Context) ginserver.Result {
		return ginserver.Data(gin.H{"id": "1"})
	})

	e := gin.New()
	err := reg.MountYAML(e, []byte(`
routes:
  - {method: GET, path: /users, do: UsersList}
  - {method: POST, path: /users, do: UserCreate}
`), nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users", nil))
	if w.Code != 200 || w.Body.String() != `{"data":["a"]}` {
		t.Fatalf("list: %d %s", w.Code, w.Body)
	}
	w = httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/users", nil))
	if w.Code != 200 || w.Body.String() != `{"data":{"id":"1"}}` {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
}

func TestPerRouteBasicAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accounts := ginserver.Accounts{"svc": "secret"}
	reg := ginroutes.NewHandlerRegistry()
	reg.RegisterHandlers(map[string]ginserver.HandlerFunc{
		"Open":   func(*gin.Context) ginserver.Result { return ginserver.Data("ok") },
		"Secure": func(*gin.Context) ginserver.Result { return ginserver.Data("sec") },
	})
	e := gin.New()
	if err := reg.MountYAML(e, []byte(`
routes:
  - {method: GET, path: /open, do: Open}
  - {method: GET, path: /secure, do: Secure, auth: [basic]}
`), accounts); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/open", nil))
	if w.Code != 200 {
		t.Fatalf("open without auth: %d", w.Code)
	}

	w = httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/secure", nil))
	if w.Code != 401 {
		t.Fatalf("secure without auth: %d", w.Code)
	}

	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("svc:secret")))
	e.ServeHTTP(w, req)
	if w.Code != 200 || w.Body.String() != `{"data":"sec"}` {
		t.Fatalf("secure with auth: %d %s", w.Code, w.Body)
	}
}

func TestBasicAuthRequiredPanics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when auth: basic and accounts empty")
		}
	}()
	reg := ginroutes.NewHandlerRegistry()
	reg.RegisterHandler("Secure", func(*gin.Context) ginserver.Result {
		return ginserver.Data("sec")
	})
	_ = reg.MountYAML(gin.New(), []byte(`
routes:
  - {method: GET, path: /secure, do: Secure, auth: [basic]}
`), nil)
}
