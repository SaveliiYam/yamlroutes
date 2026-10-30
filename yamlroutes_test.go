package yamlroutes_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/yamlroutes"
	"github.com/SaveliiYam/yamlroutes/ginroutes"
	"github.com/SaveliiYam/yamlroutes/nethttp"
)

const yml = `
routes:
  - {method: GET, path: /users, do: UsersList}
  - {method: GET, path: /users/:id, do: UserGet}
`

func TestGinAdapter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reg := ginroutes.NewRegistry().RegisterHandlers(map[string]gin.HandlerFunc{
		"UsersList": func(c *gin.Context) { c.String(200, "list") },
		"UserGet":   func(c *gin.Context) { c.String(200, "get:"+c.Param("id")) },
	})
	e := gin.New()
	if err := ginroutes.MountYAML(e, reg, []byte(yml)); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users/7", nil))
	if w.Body.String() != "get:7" {
		t.Fatalf("body = %s", w.Body)
	}
}

func TestNetHTTPAdapter(t *testing.T) {
	reg := nethttp.NewRegistry().RegisterHandlers(map[string]http.HandlerFunc{
		"UsersList": func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("list")) },
		"UserGet": func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("get:" + r.PathValue("id")))
		},
	})
	mux := http.NewServeMux()
	if err := nethttp.MountYAML(mux, reg, []byte(yml)); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users/7", nil))
	if w.Body.String() != "get:7" {
		t.Fatalf("body = %s", w.Body)
	}
}

func TestMissingHandlerAndValidation(t *testing.T) {
	reg := ginroutes.NewRegistry()
	err := ginroutes.MountYAML(gin.New(), reg, []byte("routes:\n  - {method: GET, path: /x, do: Nope}\n"))
	if err == nil || !strings.Contains(err.Error(), "Nope") {
		t.Fatalf("err = %v", err)
	}
	for _, bad := range []string{
		"routes:\n  - {method: FETCH, path: /x, do: A}\n",
		"routes:\n  - {method: GET, path: x, do: A}\n",
		"routes:\n  - {method: GET, path: /x, do: A}\n  - {method: GET, path: /x, do: B}\n",
		"routes:\n  - {method: GET, path: /x, doo: A}\n", // опечатка в ключе
		"routes:\n  - {method: GET, path: /x, do: A, auth: [oauth]}\n",
	} {
		if _, err := yamlroutes.Parse([]byte(bad)); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
	}
}

func TestAuthField(t *testing.T) {
	spec, err := yamlroutes.Parse([]byte(`
routes:
  - {method: GET, path: /open, do: Open}
  - {method: POST, path: /secure, do: Secure, auth: [basic]}
`))
	if err != nil {
		t.Fatal(err)
	}
	if spec.Routes[0].HasAuth("basic") {
		t.Fatal("open must not require basic")
	}
	if !spec.Routes[1].HasAuth("basic") || !spec.RequiresAuth("basic") {
		t.Fatal("secure must require basic")
	}
}
