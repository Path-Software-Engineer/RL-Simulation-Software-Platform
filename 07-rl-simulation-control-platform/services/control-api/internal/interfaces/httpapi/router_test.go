package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOperatorTokenComparison(t *testing.T) {
	server := &Server{operatorToken: "registered-operator-token"}

	if !server.matchesToken("registered-operator-token") {
		t.Fatal("registered operator token was rejected")
	}
	if server.matchesToken("registered-operator-taken") {
		t.Fatal("different operator token was accepted")
	}
	if server.matchesToken("short") {
		t.Fatal("different-length operator token was accepted")
	}
}

func TestPaginationParsersRejectInvalidValues(t *testing.T) {
	if _, err := parsePositiveInt("0", "limit"); err == nil {
		t.Fatal("zero page size was accepted")
	}
	if _, err := parseNonNegativeInt("-1", "afterStep", -1); err == nil {
		t.Fatal("negative explicit cursor was accepted")
	}
	if _, err := parseTimeCursor("not-a-time", "before"); err == nil {
		t.Fatal("invalid timestamp cursor was accepted")
	}
}

func TestPaginationParsersAcceptDefaults(t *testing.T) {
	limit, err := parsePositiveInt("", "limit")
	if err != nil || limit != 0 {
		t.Fatalf("default page size = %d, %v; want 0, nil", limit, err)
	}
	step, err := parseNonNegativeInt("", "afterStep", -1)
	if err != nil || step != -1 {
		t.Fatalf("default step cursor = %d, %v; want -1, nil", step, err)
	}
}

func TestPublicReadsAndOperatorCommandsHaveSeparateAuthBoundaries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{operatorToken: "registered-operator-token"}
	router := gin.New()
	router.GET("/public-evidence", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	operator := router.Group("")
	operator.Use(server.requireBearer())
	operator.POST("/operator-command", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	publicResponse := httptest.NewRecorder()
	router.ServeHTTP(
		publicResponse,
		httptest.NewRequest(http.MethodGet, "/public-evidence", nil),
	)
	if publicResponse.Code != http.StatusNoContent {
		t.Fatalf("anonymous public read = %d; want 204", publicResponse.Code)
	}

	unauthenticatedResponse := httptest.NewRecorder()
	router.ServeHTTP(
		unauthenticatedResponse,
		httptest.NewRequest(http.MethodPost, "/operator-command", nil),
	)
	if unauthenticatedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous operator command = %d; want 401", unauthenticatedResponse.Code)
	}

	authenticatedRequest := httptest.NewRequest(http.MethodPost, "/operator-command", nil)
	authenticatedRequest.Header.Set("Authorization", "Bearer registered-operator-token")
	authenticatedResponse := httptest.NewRecorder()
	router.ServeHTTP(authenticatedResponse, authenticatedRequest)
	if authenticatedResponse.Code != http.StatusNoContent {
		t.Fatalf("authenticated operator command = %d; want 204", authenticatedResponse.Code)
	}
}
