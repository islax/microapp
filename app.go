package microapp

import (
	"context"
	"database/sql"
	"net/http"

	"time"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/file"
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
		Addr:         "0.0.0.0:80",
		WriteTimeout: time.Second * time.Duration(app.Config.GetInt("HTTP_WRITE_TIMEOUT")),
		ReadTimeout:  time.Second * time.Duration(app.Config.GetInt("HTTP_READ_TIMEOUT")),
		IdleTimeout:  time.Second * time.Duration(app.Config.GetInt("HTTP_IDLE_TIMEOUT")),
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

// MigrateDB Looks for migrations directory and runs the migrations scripts in that directory
func (app *App) MigrateDB(connectionString string) {
	logger := app.log

	logger.Info("============ DB Migration Begin ============")
	fsrc, err := (&file.File{}).Open("file://migrations")
	if err != nil {
		logger.Info("No migrations directory found. Skipping migrations. Error: ", err)
		logger.Info("============ DB Migration End ============")
		return
	}
	migrateDB, err := sql.Open("mysql", connectionString)
	if err != nil {
		logger.Fatal("Unable to open DB connection for migration: ", err)
	}
	migrateDBDriver, err := mysql.WithInstance(migrateDB, &mysql.Config{})
	if err != nil {
		logger.Error("Unable to prepare DB instance for migration: ", err)
	}
	m, err := migrate.NewWithInstance("file", fsrc, "mysql", migrateDBDriver)
	if err != nil {
		logger.Error("Unable to initialize DB instance for migration: ", err)
	}
	err = m.Up()
	if err != nil {
		if err.Error() == "no change" {
			logger.Info("DB already in latest state.")
		} else {
			logger.Error("Failed to migrate DB: ", err)
			// panic(err)
		}
	} else {
		logger.Info("Successfully upgraded DB")
	}
	logger.Info("============ DB Migration End ============")
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
