package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	"github.com/heroiclabs/nakama/v3/src/app/auth"
	"github.com/heroiclabs/nakama/v3/src/app/battles"
	"github.com/heroiclabs/nakama/v3/src/app/bot"
	"github.com/heroiclabs/nakama/v3/src/app/groups"
	leaderboardsvc "github.com/heroiclabs/nakama/v3/src/app/leaderboard"
)

type ServerConfig struct {
	Logger             *zap.Logger
	AuthService        *auth.Service
	GroupService       *groups.Service
	BattleService      *battles.Service
	LeaderboardService *leaderboardsvc.Service
	BotService         *bot.Service
}

// Server wires HTTP endpoints to application services with observability instrumentation.
type Server struct {
	cfg            ServerConfig
	router         *mux.Router
	httpMetrics    *prometheus.HistogramVec
	requestCounter *prometheus.CounterVec
}

func NewServer(cfg ServerConfig) *Server {
	srv := &Server{cfg: cfg}
	srv.initMetrics()
	srv.buildRouter()
	return srv
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) initMetrics() {
	s.httpMetrics = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "sandai",
		Subsystem: "http",
		Name:      "request_latency_seconds",
		Help:      "HTTP request latency",
		Buckets:   prometheus.DefBuckets,
	}, []string{"route", "method", "code"})
	s.requestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sandai",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total HTTP requests by route",
	}, []string{"route", "method", "code"})
	prometheus.MustRegister(s.httpMetrics, s.requestCounter)
}

func (s *Server) buildRouter() {
	r := mux.NewRouter()
	r.Use(s.correlationMiddleware)
	r.Use(s.loggingMiddleware)
	r.Use(s.metricsMiddleware)

	apiRouter := r.PathPrefix("/v1").Subrouter()
	apiRouter.Handle("/auth/login", otelhttp.NewHandler(http.HandlerFunc(s.handleAuthLogin), "AuthLogin")).Methods(http.MethodPost)
	apiRouter.Handle("/groups", otelhttp.NewHandler(http.HandlerFunc(s.handleCreateGroup), "CreateGroup")).Methods(http.MethodPost)
	apiRouter.Handle("/battles", otelhttp.NewHandler(http.HandlerFunc(s.handleStartBattle), "StartBattle")).Methods(http.MethodPost)
	apiRouter.Handle("/leaderboard/{season}", otelhttp.NewHandler(http.HandlerFunc(s.handleSubmitScore), "SubmitLeaderboard")).Methods(http.MethodPost)
	apiRouter.Handle("/bot/webhook", otelhttp.NewHandler(http.HandlerFunc(s.handleBotWebhook), "BotWebhook")).Methods(http.MethodPost)

	r.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
	s.router = r
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

type errorResponse struct {
	Error string `json:"error"`
}

func (s *Server) writeError(w http.ResponseWriter, status int, err error) {
	s.writeJSON(w, status, errorResponse{Error: err.Error()})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		s.cfg.Logger.Info("http_request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rw.status),
			zap.Duration("duration", time.Since(start)),
			zap.String("request_id", correlationIDFromContext(r.Context())),
		)
	})
}

func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		route := mux.CurrentRoute(r)
		routeName := "unknown"
		if route != nil {
			if tmpl, err := route.GetPathTemplate(); err == nil {
				routeName = tmpl
			}
		}
		codeLabel := strconv.Itoa(rw.status)
		labels := prometheus.Labels{"route": routeName, "method": r.Method, "code": codeLabel}
		s.httpMetrics.With(labels).Observe(time.Since(start).Seconds())
		s.requestCounter.With(labels).Inc()
	})
}

// responseWriter captures HTTP status codes for logging/metrics.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
