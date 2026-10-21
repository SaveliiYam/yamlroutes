package ginroutes_test

import (
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
	// Чистые функции: ни одного Wrap в прикладном коде.
	reg.RegisterHandler("UsersList", func(*gin.Context) (any, error) {
		return []string{"a"}, nil
	})
	reg.RegisterHandlerStatus("UserCreate", http.StatusCreated, func(*gin.Context) (any, error) {
		return gin.H{"id": "1"}, nil
	})

	e := gin.New()
	err := reg.MountYAML(e, []byte(`
routes:
  - {method: GET, path: /users, do: UsersList}
  - {method: POST, path: /users, do: UserCreate}
`))
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
	if w.Code != 201 || w.Body.String() != `{"data":{"id":"1"}}` {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
	var _ = ginserver.Wrap // зависимость осознанная: оборачивает именно ginserver
}
