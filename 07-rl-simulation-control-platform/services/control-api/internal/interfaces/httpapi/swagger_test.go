package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSwaggerUIRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerSwaggerUI(router)

	redirect := httptest.NewRecorder()
	router.ServeHTTP(redirect, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if redirect.Code != http.StatusPermanentRedirect || redirect.Header().Get("Location") != "/docs/" {
		t.Fatalf("GET /docs = %d, %q; want 308, /docs/", redirect.Code, redirect.Header().Get("Location"))
	}

	page := httptest.NewRecorder()
	router.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/docs/", nil))
	if page.Code != http.StatusOK {
		t.Fatalf("GET /docs/ = %d; want 200", page.Code)
	}
	if !strings.Contains(page.Body.String(), `url: "/openapi.json"`) || !strings.Contains(page.Body.String(), "SwaggerUIBundle") {
		t.Fatal("Swagger UI does not load the public OpenAPI contract")
	}
	if page.Header().Get("Content-Security-Policy") == "" {
		t.Fatal("Swagger UI response is missing its content security policy")
	}
}
