package main

import (
	"evolve/config"
	"evolve/controller"
	"evolve/routes"
	"evolve/util"
	"fmt"
	"net/http"

	"github.com/rs/cors"
)

func main() {
	var logger = util.NewLogger()

	// Register routes.
	http.HandleFunc(routes.TEST, controller.Test)
	http.HandleFunc(routes.EA, controller.CreateEA)
	http.HandleFunc(routes.GP, controller.CreateGP)
	http.HandleFunc(routes.ML, controller.CreateML)
	http.HandleFunc(routes.PSO, controller.CreatePSO)
	http.HandleFunc(routes.RUNS, controller.UserRuns)
	http.HandleFunc(routes.SHARE_RUN, controller.ShareRun)

	logger.Info(fmt.Sprintf("Test http server on http://localhost%v/api/test", config.PORT))

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(http.DefaultServeMux)
	if err := http.ListenAndServe(config.PORT, corsHandler); err != nil {
		logger.Error(fmt.Sprintf("Failed to start server: %v", err))
		return
	}
}
