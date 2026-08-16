package httpapi

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Path-Software-Engineer/rl-simulation-control-platform/services/control-api/internal/application"
	"github.com/Path-Software-Engineer/rl-simulation-control-platform/services/control-api/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Server struct {
	service        *application.Service
	hub            *Hub
	allowedOrigins map[string]struct{}
	openAPIPath    string
	operatorToken  string
}

func NewRouter(service *application.Service, hub *Hub, operatorToken string, allowedOrigins []string, openAPIPath string) *gin.Engine {
	binding.EnableDecoderDisallowUnknownFields = true
	server := &Server{
		service:        service,
		hub:            hub,
		allowedOrigins: toSet(allowedOrigins),
		openAPIPath:    openAPIPath,
		operatorToken:  operatorToken,
	}
	router := gin.New()
	router.Use(gin.Recovery(), server.traceMiddleware(), server.corsMiddleware())
	router.GET("/health/live", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "alive"}) })
	router.GET("/health/ready", server.ready)
	router.StaticFile("/openapi.json", openAPIPath)

	api := router.Group("/api/v1")
	api.Use(server.requireBearer())
	api.GET("/environments", server.listEnvironments)
	api.GET("/environments/:id", server.getEnvironment)
	api.GET("/policies", server.listPolicies)
	api.GET("/policies/:id", server.getPolicy)
	api.POST("/training-runs", server.createRun)
	api.GET("/training-runs", server.listRuns)
	api.GET("/training-runs/:id", server.getRun)
	api.POST("/training-runs/:id/pause", server.controlRun("pause"))
	api.POST("/training-runs/:id/resume", server.controlRun("resume"))
	api.POST("/training-runs/:id/cancel", server.controlRun("cancel"))
	api.GET("/training-runs/:id/metrics", server.listMetrics)
	api.GET("/training-runs/:id/episodes", server.listEpisodes)
	api.GET("/episodes/:id", server.getEpisode)
	api.GET("/episodes/:id/transitions", server.listTransitions)
	api.POST("/episodes/:id/feedback", server.createFeedback)
	router.GET("/ws/v1/runs/:id", server.subscribeRun)
	router.NoRoute(func(c *gin.Context) { server.problem(c, domain.ErrNotFound) })
	return router
}

func (server *Server) ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := server.service.Ready(ctx); err != nil {
		slog.Warn("readiness dependency unavailable", "trace_id", traceID(c), "error", err)
		server.problemWithDetail(c, http.StatusServiceUnavailable, "dependency_unavailable", "A required dependency is unavailable.", "")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready", "database": "reachable", "stream": "reachable"})
}

func (server *Server) listEnvironments(c *gin.Context) {
	items, err := server.service.ListEnvironments(c.Request.Context())
	server.respond(c, items, err)
}

func (server *Server) getEnvironment(c *gin.Context) {
	item, err := server.service.GetEnvironment(c.Request.Context(), c.Param("id"))
	server.respond(c, item, err)
}

func (server *Server) listPolicies(c *gin.Context) {
	items, err := server.service.ListPolicies(c.Request.Context())
	server.respond(c, items, err)
}

func (server *Server) getPolicy(c *gin.Context) {
	item, err := server.service.GetPolicy(c.Request.Context(), c.Param("id"))
	server.respond(c, item, err)
}

func (server *Server) createRun(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	var command domain.CreateRunCommand
	if err := c.ShouldBindJSON(&command); err != nil {
		server.problemWithDetail(c, http.StatusBadRequest, "invalid_request", "The run request is invalid.", err.Error())
		return
	}
	run, err := server.service.CreateRun(c.Request.Context(), command, c.GetHeader("Idempotency-Key"), actor(c))
	if err != nil {
		server.problem(c, err)
		return
	}
	c.JSON(http.StatusAccepted, run)
}

func (server *Server) listRuns(c *gin.Context) {
	query, err := parseRunQuery(c)
	if err != nil {
		server.problem(c, err)
		return
	}
	items, err := server.service.ListRuns(c.Request.Context(), query)
	if err != nil {
		server.problem(c, err)
		return
	}
	limit := effectiveLimit(query.Limit, 50, 200)
	if len(items) == limit {
		c.Header("X-Next-Cursor", items[len(items)-1].CreatedAt.UTC().Format(time.RFC3339Nano))
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit})
}

func (server *Server) getRun(c *gin.Context) {
	item, err := server.service.GetRun(c.Request.Context(), c.Param("id"))
	server.respond(c, item, err)
}

func (server *Server) controlRun(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		run, err := server.service.RequestControl(c.Request.Context(), c.Param("id"), action, c.GetHeader("Idempotency-Key"), actor(c))
		if err != nil {
			server.problem(c, err)
			return
		}
		c.JSON(http.StatusAccepted, run)
	}
}

func (server *Server) listMetrics(c *gin.Context) {
	query, err := parseMetricQuery(c)
	if err != nil {
		server.problem(c, err)
		return
	}
	items, err := server.service.ListRunMetrics(c.Request.Context(), c.Param("id"), query)
	if len(items) == effectiveLimit(query.Limit, 50, 200) {
		c.Header("X-Next-Cursor", items[len(items)-1].SampledAt.UTC().Format(time.RFC3339Nano))
	}
	server.respond(c, items, err)
}

func (server *Server) listEpisodes(c *gin.Context) {
	query, err := parseEpisodeQuery(c)
	if err != nil {
		server.problem(c, err)
		return
	}
	items, err := server.service.ListRunEpisodes(c.Request.Context(), c.Param("id"), query)
	if len(items) == effectiveLimit(query.Limit, 50, 100) {
		c.Header("X-Next-Cursor", strconv.Itoa(items[len(items)-1].EpisodeNumber))
	}
	server.respond(c, items, err)
}

func (server *Server) getEpisode(c *gin.Context) {
	item, err := server.service.GetEpisode(c.Request.Context(), c.Param("id"))
	server.respond(c, item, err)
}

func (server *Server) listTransitions(c *gin.Context) {
	query, err := parseTransitionQuery(c)
	if err != nil {
		server.problem(c, err)
		return
	}
	items, err := server.service.ListTransitions(c.Request.Context(), c.Param("id"), query)
	if len(items) == effectiveLimit(query.Limit, 100, 200) {
		c.Header("X-Next-Cursor", strconv.Itoa(items[len(items)-1].StepIndex))
	}
	server.respond(c, items, err)
}

func (server *Server) createFeedback(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10)
	var command domain.FeedbackCommand
	if err := c.ShouldBindJSON(&command); err != nil {
		server.problemWithDetail(c, http.StatusBadRequest, "invalid_feedback", "The feedback is invalid.", err.Error())
		return
	}
	if err := server.service.CreateFeedback(c.Request.Context(), c.Param("id"), c.GetHeader("Idempotency-Key"), actor(c), command); err != nil {
		server.problem(c, err)
		return
	}
	c.Status(http.StatusCreated)
}

func (server *Server) subscribeRun(c *gin.Context) {
	runID := c.Param("id")
	if _, err := uuid.Parse(runID); err != nil {
		server.problem(c, domain.ErrInvalidArgument)
		return
	}
	protocols := websocket.Subprotocols(c.Request)
	if len(protocols) != 2 || protocols[0] != "rl-run-v1" || !server.matchesToken(protocols[1]) {
		server.problemWithDetail(c, http.StatusUnauthorized, "unauthorized", "Operator authentication is required.", "")
		return
	}
	upgrader := websocket.Upgrader{
		HandshakeTimeout: 5 * time.Second,
		Subprotocols:     []string{"rl-run-v1"},
		CheckOrigin: func(request *http.Request) bool {
			_, allowed := server.allowedOrigins[request.Header.Get("Origin")]
			return allowed
		},
	}
	connection, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer connection.Close()
	updates, unsubscribe := server.hub.Subscribe(runID)
	defer unsubscribe()
	connection.SetReadLimit(1024)
	_ = connection.SetReadDeadline(time.Now().Add(45 * time.Second))
	connection.SetPongHandler(func(string) error {
		return connection.SetReadDeadline(time.Now().Add(45 * time.Second))
	})
	go func() {
		for {
			if _, _, readErr := connection.ReadMessage(); readErr != nil {
				_ = connection.Close()
				return
			}
		}
	}()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case payload, ok := <-updates:
			if !ok || connection.WriteMessage(websocket.TextMessage, payload) != nil {
				return
			}
		case <-time.After(20 * time.Second):
			if connection.WriteControl(websocket.PingMessage, nil, time.Now().Add(2*time.Second)) != nil {
				return
			}
		}
	}
}

func (server *Server) respond(c *gin.Context, value any, err error) {
	if err != nil {
		server.problem(c, err)
		return
	}
	c.JSON(http.StatusOK, value)
}

func (server *Server) problem(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		server.problemWithDetail(c, http.StatusNotFound, "not_found", "The requested resource was not found.", "")
	case errors.Is(err, domain.ErrInvalidState):
		server.problemWithDetail(c, http.StatusConflict, "invalid_run_state", "The command is not valid for the current run state.", "")
	case errors.Is(err, domain.ErrConflict):
		server.problemWithDetail(c, http.StatusConflict, "idempotency_conflict", "The idempotency key was already used for a different request.", "")
	case errors.Is(err, domain.ErrInvalidArgument):
		server.problemWithDetail(c, http.StatusBadRequest, "invalid_argument", "One or more request values are invalid.", "")
	default:
		server.problemWithDetail(c, http.StatusInternalServerError, "internal_error", "The platform could not complete the request.", "")
	}
}

func (server *Server) problemWithDetail(c *gin.Context, status int, code, title, detail string) {
	c.Header("Content-Type", "application/problem+json")
	c.JSON(status, gin.H{
		"type":    "https://paths.dev/problems/" + code,
		"title":   title,
		"status":  status,
		"detail":  detail,
		"code":    code,
		"traceId": traceID(c),
	})
}

func (server *Server) traceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		trace := c.GetHeader("X-Request-ID")
		if _, err := uuid.Parse(trace); err != nil {
			trace = uuid.NewString()
		}
		c.Set("trace_id", trace)
		c.Header("X-Request-ID", trace)
		c.Next()
	}
}

func (server *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, allowed := server.allowedOrigins[origin]; allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Request-ID")
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Header("Access-Control-Expose-Headers", "X-Next-Cursor, X-Request-ID")
		}
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (server *Server) requireBearer() gin.HandlerFunc {
	return func(c *gin.Context) {
		const prefix = "Bearer "
		authorization := c.GetHeader("Authorization")
		if !strings.HasPrefix(authorization, prefix) || !server.matchesToken(strings.TrimPrefix(authorization, prefix)) {
			server.problemWithDetail(c, http.StatusUnauthorized, "unauthorized", "Operator authentication is required.", "")
			c.Abort()
			return
		}
		c.Next()
	}
}

func (server *Server) matchesToken(candidate string) bool {
	if len(candidate) != len(server.operatorToken) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(candidate), []byte(server.operatorToken)) == 1
}

func actor(c *gin.Context) string { return "local-operator" }

func traceID(c *gin.Context) string {
	value, _ := c.Get("trace_id")
	return fmt.Sprint(value)
}

func parseRunQuery(c *gin.Context) (domain.RunQuery, error) {
	limit, err := parsePositiveInt(c.Query("limit"), "limit")
	if err != nil {
		return domain.RunQuery{}, err
	}
	before, err := parseTimeCursor(c.Query("before"), "before")
	if err != nil {
		return domain.RunQuery{}, err
	}
	return domain.RunQuery{Limit: limit, Status: domain.RunStatus(c.Query("status")), Before: before}, nil
}

func parseMetricQuery(c *gin.Context) (domain.MetricQuery, error) {
	limit, err := parsePositiveInt(c.Query("limit"), "limit")
	if err != nil {
		return domain.MetricQuery{}, err
	}
	after, err := parseTimeCursor(c.Query("after"), "after")
	if err != nil {
		return domain.MetricQuery{}, err
	}
	return domain.MetricQuery{Limit: limit, Metric: c.Query("metric"), After: after}, nil
}

func parseEpisodeQuery(c *gin.Context) (domain.EpisodeQuery, error) {
	limit, err := parsePositiveInt(c.Query("limit"), "limit")
	if err != nil {
		return domain.EpisodeQuery{}, err
	}
	afterEpisode, err := parseNonNegativeInt(c.Query("afterEpisode"), "afterEpisode", 0)
	if err != nil {
		return domain.EpisodeQuery{}, err
	}
	return domain.EpisodeQuery{Limit: limit, AfterEpisode: afterEpisode}, nil
}

func parseTransitionQuery(c *gin.Context) (domain.TransitionQuery, error) {
	limit, err := parsePositiveInt(c.Query("limit"), "limit")
	if err != nil {
		return domain.TransitionQuery{}, err
	}
	afterStep, err := parseNonNegativeInt(c.Query("afterStep"), "afterStep", -1)
	if err != nil {
		return domain.TransitionQuery{}, err
	}
	return domain.TransitionQuery{Limit: limit, AfterStep: afterStep}, nil
}

func parsePositiveInt(raw, field string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("%w: %s", domain.ErrInvalidArgument, field)
	}
	return value, nil
}

func parseNonNegativeInt(raw, field string, fallback int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%w: %s", domain.ErrInvalidArgument, field)
	}
	return value, nil
}

func parseTimeCursor(raw, field string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	value, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrInvalidArgument, field)
	}
	value = value.UTC()
	return &value, nil
}

func effectiveLimit(value, fallback, maximum int) int {
	if value <= 0 {
		return fallback
	}
	if value > maximum {
		return maximum
	}
	return value
}

func toSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result[trimmed] = struct{}{}
		}
	}
	return result
}
