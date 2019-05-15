package microapp

import (
	"context"
	"net/http"

	"time"

	"github.com/gorilla/mux"
	"github.com/islax/microapp/config"
	"github.com/islax/microapp/event"
	"github.com/islax/microapp/repository"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// RouteSpecifier should be implemented by the class that sets routes for the API endpoints
type RouteSpecifier interface {
	RegisterRoutes(router *mux.Router)
}

// App structure for tenant microservice
type App struct {
	Name            string
	Config          *config.Config
	DB              *gorm.DB
	Router          *mux.Router
	server          *http.Server
	log             *log.Logger
	eventDispatcher event.EventDispatcher
}

// New creates a new microApp
func New(appName string, appConfigOverride map[string]string, appLog *log.Logger, appDB *gorm.DB, appEventDispatcher event.EventDispatcher) *App {
	appConfig := config.NewConfig(appConfigOverride)
	return &App{Name: appName, Config: appConfig, log: appLog, DB: appDB, eventDispatcher: appEventDispatcher}
}

// NewUnitOfWork creates new UnitOfWork
func (app *App) NewUnitOfWork(readOnly bool) *repository.UnitOfWork {
	return repository.NewUnitOfWork(app.DB, readOnly)
}

//Initialize initializes properties of the app
func (app *App) Initialize(routeSpecifiers []RouteSpecifier) {
	app.Router = mux.NewRouter()
	app.Router.Use(mux.CORSMethodMiddleware(app.Router))
	app.Router.Use(app.loggingMiddleware)

	for _, routeSpecifier := range routeSpecifiers {
		routeSpecifier.RegisterRoutes(app.Router)
	}

	app.server = &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      app.Router,
	}
}

//Start http server and start listening to the requests
func (app *App) Start() {
	if err := app.server.ListenAndServe(); err != nil {
		log.Println(err)
	} else {
		log.Println("Server started")
	}
}

// Logger returns logger for specified module
func (app *App) Logger(module string) *log.Entry {
	return app.log.WithFields(
		log.Fields{
			"service": app.Name,
			"module":  module,
		})
}

// Stop http server
func (app *App) Stop() {
	wait, _ := time.ParseDuration("2m")
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	app.server.Shutdown(ctx)
}

func (app *App) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.Logger("HTTP").Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

// DispatchEvent delegates to eventDispatcher.
func (app *App) DispatchEvent(token string, topic string, payload interface{}) {
	if app.eventDispatcher != nil {
		app.eventDispatcher.DispatchEvent(token, topic, payload)
	}
}
