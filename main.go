package main

import (
	"context"
	"errors"
	"evolve/controller"
	"evolve/modules/sse"
	"evolve/modules/ws"
	"evolve/routes"
	"evolve/util"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/cors"
)

var (
	PORT         string
	FRONTEND_URL string
)

func main() {
	PORT = fmt.Sprintf(":%s", os.Getenv("HTTP_PORT"))
	if PORT == ":" {
		PORT = ":5002"
	}
	FRONTEND_URL = os.Getenv("FRONTEND_URL")
	if FRONTEND_URL == "" {
		FRONTEND_URL = "http://localhost:3000"
	}

	var logger = util.NewLogger()

	// ---- Context and Graceful Shutdown Setup ----
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ---- Start WebSocket Server ----
	wsHub := ws.StartServer(ctx, *logger)
	logger.Info("WebSocket Hub started.")

	// ---- Register HTTP Routes ----
	mux := http.DefaultServeMux

	mux.HandleFunc(routes.TEST, controller.Test)
	mux.HandleFunc(routes.EA, controller.CreateEA)
	mux.HandleFunc(routes.GP, controller.CreateGP)
	mux.HandleFunc(routes.ML, controller.CreateML)
	mux.HandleFunc(routes.PSO, controller.CreatePSO)
	mux.HandleFunc(routes.RUNS, controller.UserRuns)
	mux.HandleFunc(routes.SHARE_RUN, controller.ShareRun)
	mux.HandleFunc(routes.RUN, controller.UserRun)

	// WebSocket Route
	wsHandler := ws.GetHandler(ctx, wsHub, *logger)
	mux.HandleFunc(routes.LIVE, wsHandler)
	logger.Info("WebSocket endpoint registered at /live/")

	// ---- SSE Route ----
	sseHandler := sse.GetSSEHandler(*logger)
	mux.HandleFunc(routes.LOGS, sseHandler)
	logger.Info("SSE endpoint registered at /runs/logs/")

	// ---- Log the test endpoint URL ----
	logger.Info(fmt.Sprintf("Test http server on http://localhost%s/api/test", PORT))

	// ---- CORS Configuration ----
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{FRONTEND_URL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Ensure GET is allowed for SSE
		AllowedHeaders:   []string{"*"},                                       // Allow X-RUN-ID header
		AllowCredentials: true,
	}).Handler(mux)

	// ---- HTTP Server Setup and Start ----
	server := &http.Server{
		Addr:         PORT,
		Handler:      corsHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // Set WriteTimeout to 0 for long-lived SSE connections
		IdleTimeout:  0, // Set IdleTimeout to 0 for long-lived SSE connections
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
	}

	// Start server in a goroutine
	go func() {
		logger.Info(fmt.Sprintf("HTTP server starting on %s (Frontend: %s)", server.Addr, FRONTEND_URL))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(fmt.Sprintf("HTTP server ListenAndServe error: %v", err))
			stop()
		}
	}()

	// ---- Wait for Shutdown Signal ----
	<-ctx.Done()

	// ---- Initiate Graceful Shutdown ----
	logger.Info("Shutdown signal received. Starting graceful shutdown...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error(fmt.Sprintf("HTTP server graceful shutdown failed: %v", err))
	} else {
		logger.Info("HTTP server shutdown complete.")
	}

	logger.Info("Server exiting.")
}
