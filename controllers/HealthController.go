package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
)

// HealthController provides method to check health and readiness
type HealthController struct {
}

// NewHealthController returns a new instance of HealthController
func NewHealthController() *HealthController {
	return &HealthController{}
}

// RegisterRoutes implements interface RouteSpecifier
func (controller *HealthController) RegisterRoutes(router *mux.Router) {
	healthRouter := router.PathPrefix("/health").Subrouter()
	healthRouter.HandleFunc("", controller.healthCheck).Methods("GET")
}

func (controller *HealthController) healthCheck(w http.ResponseWriter, r *http.Request) {
}
