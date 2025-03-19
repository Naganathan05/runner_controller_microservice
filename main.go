package main

import (
	// "evolve/config"
	"evolve/controller"
	"evolve/routes"
	"evolve/util"
	"fmt"
	"github.com/rs/cors"
	"net/http"
	"os"
)

var (
	PORT         string
	FRONTEND_URL string
)

func main() {

	PORT = fmt.Sprintf(":%s", os.Getenv("HTTP_PORT"))
	FRONTEND_URL = os.Getenv("FRONTEND_URL")

	var logger = util.NewLogger()

	// Register routes.
	http.HandleFunc(routes.TEST, controller.Test)
	http.HandleFunc(routes.EA, controller.CreateEA)
	http.HandleFunc(routes.GP, controller.CreateGP)
	http.HandleFunc(routes.ML, controller.CreateML)
	http.HandleFunc(routes.PSO, controller.CreatePSO)
	http.HandleFunc(routes.RUNS, controller.UserRuns)
	http.HandleFunc(routes.SHARE_RUN, controller.ShareRun)
	http.HandleFunc(routes.RUN, controller.UserRun)

	logger.Info(fmt.Sprintf("Test http server on http://localhost%v/api/test", PORT))

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{FRONTEND_URL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(http.DefaultServeMux)
	if err := http.ListenAndServe("0.0.0.0"+PORT, corsHandler); err != nil {
		logger.Error(fmt.Sprintf("Failed to start server: %v", err))
		return
	}
}
