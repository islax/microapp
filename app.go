package microapp

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"gorm.io/gorm/schema"

	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/islax/microapp/config"
	microappCtx "github.com/islax/microapp/context"
	"github.com/islax/microapp/event"
	"github.com/islax/microapp/log"
	"github.com/islax/microapp/metrics"
	"github.com/islax/microapp/repository"
	"github.com/islax/microapp/retry"
	"github.com/islax/microapp/security"
	gormmysqldriver "gorm.io/driver/mysql"
	"gorm.io/gorm"

	gomysqldriver "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
	prometheusmetrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
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
	MemcachedClient *memcache.Client
	Router          *mux.Router
	server          *http.Server
	log             zerolog.Logger
	eventDispatcher event.Dispatcher
}

// NewWithEnvValues creates a new application with environment variable values for initializing database, event dispatcher and logger.
func NewWithEnvValues(appName string, appConfigDefaults map[string]interface{}) *App {
	appConfig := config.NewConfig(appConfigDefaults)
	printMicroAppVersion(appConfig)
	log.InitializeGlobalSettings()
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	consoleOnlyLogger := log.New(appName, appConfig.GetString("LOG_LEVEL"), os.Stdout)
	multiWriters := io.MultiWriter(os.Stdout)
	//To log a human-friendly, colorized output
	if appConfig.GetString("FORMAT_CONSOLE_LOG") == "true" {
		consoleOnlyLogger = log.New(appName, appConfig.GetString("LOG_LEVEL"), consoleWriter)
		multiWriters = io.MultiWriter(consoleWriter)
	}
	consoleOnlyLogger.Info().Msgf("Staring: %v", appName)
	// consoleOnlyLogger := zerolog.New(consoleWriter).With().Timestamp().Str("service", appName).Logger().Level()

	var err error
	var appEventDispatcher event.Dispatcher
	if appConfig.GetStringWithDefault("ENABLE_EVENT_DISPATCHER", "0") == "1" || appConfig.GetStringWithDefault("LOG_TO_EVENTQ", "0") == "1" {
		if appEventDispatcher, err = event.NewRabbitMQEventDispatcher(consoleOnlyLogger); err != nil {
			consoleOnlyLogger.Fatal().Err(err).Msg("Failed to initialize event dispatcher to queue, exiting the application!")
		}
		if appConfig.GetStringWithDefault("LOG_TO_EVENTQ", "0") == "1" {
			multiWriters = io.MultiWriter(os.Stdout, event.NewEventQWriter(appEventDispatcher))
			if appConfig.GetString("FORMAT_CONSOLE_LOG") == "true" {
				multiWriters = io.MultiWriter(consoleWriter, event.NewEventQWriter(appEventDispatcher))
			}
		}
	} else {
		consoleOnlyLogger.Warn().Msg("Event dispatcher not enabled. Please set ISLA_ENABLE_EVENT_DISPATCHER or ISLA_LOG_TO_EVENTQ to '1' to enable it.")
	}
	//TODO: default module to system
	appLogger := log.New(appName, appConfig.GetString("LOG_LEVEL"), multiWriters)
	//TODO: Need to wait till eventDispatcher is ready
	time.Sleep(5 * time.Second)

	app := App{Name: appName, Config: appConfig, log: *appLogger, eventDispatcher: appEventDispatcher}
	err = app.initializeDB()
	if err != nil {
		consoleOnlyLogger.Fatal().Err(err).Msg("Failed to initialize database, exiting the application!!")
	}

	err = app.initializeMemcache()
	if err != nil {
		consoleOnlyLogger.Fatal().Err(err).Msg("Failed to initialize memcached, exiting the application!!")
	}

	return &app
}

// New creates a new microApp
func New(appName string, appConfigDefaults map[string]interface{}, appLog zerolog.Logger, appDB *gorm.DB, appMemcache *memcache.Client, appEventDispatcher event.Dispatcher) *App {
	appConfig := config.NewConfig(appConfigDefaults)
	return &App{Name: appName, Config: appConfig, log: appLog, DB: appDB, MemcachedClient: appMemcache, eventDispatcher: appEventDispatcher}
}

func (app *App) initializeDB() error {
	if app.Config.GetBool(config.EvSuffixForDBRequired) {
		var db *gorm.DB
		err := retry.Do(3, time.Second*15, func() error {
			//gorm custom logger
			dbLogger := log.NewGormLogger(app.log, log.Config{SlowThreshold: time.Duration(app.Config.GetInt(config.EvSuffixForGormSlowThreshold)) * time.Millisecond})
			var err error
			dbconf := &gorm.Config{PrepareStmt: true, Logger: dbLogger}

			if app.Config.GetBool("DB_NAMING_STRATEGY_IS_SINGULAR") {
				dbconf.NamingStrategy = schema.NamingStrategy{SingularTable: true}
			}

			if err = registerTLSconfig(app.Config.GetString("DB_SSL_CA_PATH"), app.Config.GetString("DB_SSL_CERT_PATH"), app.Config.GetString("DB_SSL_KEY_PATH")); err != nil {
				app.log.Warn().Err(err).Msgf("TLS config error [%v]. Connecting without certificates", err)
			}

			sqlDB, err := sql.Open("mysql", app.GetConnectionString())
			if err != nil {
				app.log.Error().Err(err).Msgf("Error creating connection pool [%v]. Trying again...", err)
			}
			sqlDB.SetConnMaxLifetime(time.Duration(app.Config.GetInt(config.EvSuffixForDBConnectionLifetime)) * time.Minute)
			sqlDB.SetMaxIdleConns(app.Config.GetInt(config.EvSuffixForDBMaxIdleConnections))
			sqlDB.SetMaxOpenConns(app.Config.GetInt(config.EvSuffixForDBMaxOpenConnections))
			db, err = gorm.Open(gormmysqldriver.New(gormmysqldriver.Config{
				Conn: sqlDB,
			}), dbconf)
			if err != nil && strings.Contains(err.Error(), "connection refused") {
				app.log.Warn().Msgf("Error connecting to Database [%v]. Trying again...", err)
				return err
			}

			return retry.Stop{OriginalError: err}
		})
		app.DB = db
		app.log.Info().Msg("Database connected!")
		return err
	}
	return nil
}

// GetConnectionString gets database connection string
func (app *App) GetConnectionString() string {
	dbHost := app.Config.GetString("DB_HOST")
	dbName := app.Config.GetString("DB_NAME")
	dbPort := app.Config.GetString("DB_PORT")
	dbUser := app.Config.GetString("DB_USER")
	dbPassword := app.Config.GetString("DB_PWD")

	return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?multiStatements=true&charset=utf8&parseTime=True&loc=Local&tls=preferred", dbUser, dbPassword, dbHost, dbPort, dbName)
}

// NewUnitOfWork creates new UnitOfWork
func (app *App) NewUnitOfWork(readOnly bool, logger zerolog.Logger) *repository.UnitOfWork {
	return repository.NewUnitOfWork(app.DB, readOnly, logger, log.Config{SlowThreshold: time.Duration(app.Config.GetInt(config.EvSuffixForGormSlowThreshold)) * time.Millisecond})
}

//Initialize initializes properties of the app
func (app *App) Initialize(routeSpecifiers []RouteSpecifier) {

	logger := app.log
	app.Router = mux.NewRouter()
	app.Router.Use(mux.CORSMethodMiddleware(app.Router))

	for _, routeSpecifier := range routeSpecifiers {
		routeSpecifier.RegisterRoutes(app.Router)
	}

	//prometheus
	if app.Config.GetBool(config.EvSuffixForEnableMetrics) {
		if app.Config.GetBool(config.EvSuffixForDBRequired) {
			metrics.RegisterGormMetrics(app.DB, app.Config)
		}
		// Create our middleware.
		mdlw := middleware.New(middleware.Config{
			Recorder: prometheusmetrics.NewRecorder(prometheusmetrics.Config{}),
			Service:  app.Name,
		})
		app.Router.Use(std.HandlerProvider("", mdlw))
		app.Router.Path("/metrics").Handler(promhttp.Handler())
	}

	app.Router.Use(app.loggingMiddleware)

	//TODO: Revisit this logic
	apiPort := "80"
	if app.Config.IsSet("API_PORT") {
		port := app.Config.GetString("API_PORT")
		if _, err := strconv.Atoi(port); err != nil {
			logger.Error().Msg("API port needs to be a number. " + port + " is not a number.")
		} else {
			apiPort = port
		}
	}

	logger.Debug().Str("appname", app.Name).Msg("Api server will start on port: " + apiPort)
	app.server = &http.Server{
		Addr:         "0.0.0.0:" + apiPort,
		WriteTimeout: time.Second * time.Duration(app.Config.GetInt("HTTP_WRITE_TIMEOUT")),
		ReadTimeout:  time.Second * time.Duration(app.Config.GetInt("HTTP_READ_TIMEOUT")),
		IdleTimeout:  time.Second * time.Duration(app.Config.GetInt("HTTP_IDLE_TIMEOUT")),
		Handler:      app.Router,
	}
}

//Start http server and start listening to the requests
func (app *App) Start() {

	if app.Config.GetString("ENABLE_TLS") == "true" {
		app.StartSecure(app.Config.GetString("TLS_CRT"), app.Config.GetString("TLS_KEY"))
	} else {
		if err := app.server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				app.log.Fatal().Err(err).Msg("Unable to start server, exiting the application!")
			}
		}
	}
}

func printMicroAppVersion(c *config.Config) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Printf("Failed to read build info")
		return
	}

	for _, dep := range bi.Deps {
		if strings.Contains(dep.Path, "microapp") {
			fmt.Println("Microapp Version:" + dep.Version)
			break
		}
	}
}

//StartSecure starts https server and listens to the requests
func (app *App) StartSecure(tlsCert string, tlsKey string) {

	if tlsCert == "" {
		app.log.Fatal().Msg("TLS_CRT is not defined or empty, exiting the application!")
	}

	if tlsKey == "" {
		app.log.Fatal().Msg("TLS_KEY is not defined or empty, exiting the application!")
	}

	if err := app.server.ListenAndServeTLS(tlsCert, tlsKey); err != nil {
		app.log.Fatal().Err(err).Msg("Unable to start server or server stopped, exiting the application!")
	}
}

// Logger returns logger for specified module
func (app *App) Logger(module string) *zerolog.Logger {
	logger := app.log.With().Str("service", app.Name).Str("module", module).Logger()
	return &logger
}

// MigrateDB Looks for migrations directory and runs the migrations scripts in that directory
func (app *App) MigrateDB() {
	logger := app.log

	logger.Debug().Msg("DB Migration Begin...")
	fsrc, err := (&file.File{}).Open("file://migrations")
	if err != nil {
		logger.Info().Err(err).Msg("No migrations directory found, skipping migrations!")
		logger.Info().Msg("DB Migration End!")
		return
	}
	migrateDB, err := sql.Open("mysql", app.GetConnectionString())
	if err != nil {
		logger.Fatal().Err(err).Msg("Unable to open DB connection for migration, exiting the application!")
	}
	migrateDBDriver, err := mysql.WithInstance(migrateDB, &mysql.Config{})
	if err != nil {
		logger.Fatal().Err(err).Msg("Unable to prepare DB instance for migration, exiting the application!")
	}
	m, err := migrate.NewWithInstance("file", fsrc, "mysql", migrateDBDriver)
	if err != nil {
		logger.Fatal().Err(err).Msg("Unable to initialize DB instance for migration, exiting the application!")
	}
	err = m.Up()
	if err != nil {
		if err.Error() == "no change" {
			logger.Info().Msg("DB already in latest state.")
		} else {
			logger.Fatal().Err(err).Msg("Failed to migrate DB, exiting the application!")
		}
	} else {
		logger.Debug().Msg("Successfully upgraded DB")
	}
	logger.Info().Msg("DB Migration End!")
}

// Stop http server
func (app *App) Stop() {
	wait, _ := time.ParseDuration("2m")
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	app.server.Shutdown(ctx)

	if app.Config.GetBool("DB_REQUIRED") {
		sqlDB, err := app.DB.DB()
		if err != nil {
			sqlDB.Close()
		}
	}
}

type httpStatusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *httpStatusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func (app *App) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		if r.Header.Get("X-Correlation-ID") == "" {
			r.Header.Set("X-Correlation-ID", uuid.NewV4().String())
		}
		logger := app.Logger("Ingress").With().Timestamp().Str("caller", r.Header.Get("X-Client")).Str("correlationId", r.Header.Get("X-Correlation-ID")).Str("method", r.Method).Str("requestURI", r.RequestURI).Logger()

		rec := &httpStatusRecorder{ResponseWriter: w}
		if (!strings.HasSuffix(r.RequestURI, "/health") || app.Config.GetBool(config.EvSuffixForEnableHealthLog)) && !strings.HasSuffix(r.RequestURI, "/metrics") {
			logger.Info().Msg("Begin")
		}
		next.ServeHTTP(rec, r)
		if (!strings.HasSuffix(r.RequestURI, "/health") || app.Config.GetBool(config.EvSuffixForEnableHealthLog)) && !strings.HasSuffix(r.RequestURI, "/metrics") {
			if rec.status >= http.StatusInternalServerError {
				logger.Error().Int("status", rec.status).Dur("responseTime", time.Now().Sub(startTime)).Msg("End.")
			} else {
				logger.Info().Int("status", rec.status).Dur("responseTime", time.Now().Sub(startTime)).Msg("End.")
			}
		}
	})
}

// DispatchEvent delegates to eventDispatcher.
func (app *App) DispatchEvent(token string, corelationID string, topic string, payload interface{}) {
	if app.eventDispatcher != nil {
		app.eventDispatcher.DispatchEvent(token, corelationID, topic, payload)
	}
}

// NewExecutionContext creates new exectuion context
func (app *App) NewExecutionContext(token *security.JwtToken, correlationID string, action string, isUOWReqd, isUOWReadonly bool) microappCtx.ExecutionContext {
	executionContext := microappCtx.NewExecutionContext(token, correlationID, action, app.log)
	if isUOWReqd {
		uow := app.NewUnitOfWork(isUOWReadonly, *executionContext.GetDefaultLogger())
		executionContext.SetUOW(uow)
	}
	return executionContext
}

// NewExecutionContextWithCustomToken creates new exectuion context with custom made token
func (app *App) NewExecutionContextWithCustomToken(tenantID uuid.UUID, userID uuid.UUID, username string, correlationID string, action string, admin, isUOWReqd, isUOWReadonly bool) microappCtx.ExecutionContext {
	executionContext := microappCtx.NewExecutionContext(&security.JwtToken{Admin: admin, TenantID: tenantID, UserID: userID, UserName: username}, correlationID, action, app.log)
	if isUOWReqd {
		uow := app.NewUnitOfWork(isUOWReadonly, *executionContext.GetDefaultLogger())
		executionContext.SetUOW(uow)
	}
	return executionContext
}

// NewExecutionContextWithSystemToken creates new exectuion context with sys default token
func (app *App) NewExecutionContextWithSystemToken(correlationID string, action string, admin, isUOWReqd, isUOWReadonly bool) microappCtx.ExecutionContext {
	executionContext := microappCtx.NewExecutionContext(&security.JwtToken{Admin: admin, TenantID: uuid.Nil, UserID: uuid.Nil, TenantName: "None", UserName: "System", DisplayName: "System"}, correlationID, action, app.log)
	if isUOWReqd {
		uow := app.NewUnitOfWork(isUOWReadonly, *executionContext.GetDefaultLogger())
		executionContext.SetUOW(uow)
	}
	return executionContext
}

// GetCorrelationIDFromRequest returns correlationId from request header
func GetCorrelationIDFromRequest(r *http.Request) string {
	return r.Header.Get("X-Correlation-ID")
}

func registerTLSconfig(ssl_ca, ssl_cert, ssl_key string) error {
	rootCertPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile(ssl_ca)
	if err != nil {
		return err
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		return err
	}
	clientCert := make([]tls.Certificate, 0, 1)
	certs, err := tls.LoadX509KeyPair(ssl_cert, ssl_key)
	if err != nil {
		return err
	}
	clientCert = append(clientCert, certs)
	gomysqldriver.RegisterTLSConfig("custom", &tls.Config{
		RootCAs:      rootCertPool,
		Certificates: clientCert,
	})
	return nil
}

// initializeMemcache initializes the memcached client
func (app *App) initializeMemcache() error {
	if !app.Config.GetBool(config.EvSuffixForMemCachedRequired) {
		return nil
	}

	memcachedHost := app.Config.GetString(config.EvSuffixForMemCachedHost)
	memcachedPort := app.Config.GetString(config.EvSuffixForMemCachedPort)

	app.log.Debug().Msgf("connecting to %s\n",net.JoinHostPort(memcachedHost, memcachedPort))

	memcachedClient := memcache.New(net.JoinHostPort(memcachedHost, memcachedPort))
	if memcachedClient == nil {
		return errors.New("can not able to connect memcached client")
	}

	if err := memcachedClient.Ping(); err != nil {
		return errors.New(fmt.Sprintf("can not able to connect memcached client with err: %s", err.Error()))
	}

	// setting dummy value just to verify if connection established with memcached server and server is available
	if err := memcachedClient.Set(&memcache.Item{
		Key:        "foo",
		Value:      []byte("bar"),
		Expiration: int32(time.Now().Add(2 * time.Second).Unix()),
	}); err != nil {
		return errors.New(fmt.Sprintf("can not able to connect memcached client with err: %s", err.Error()))
	}

	app.MemcachedClient = memcachedClient

	app.log.Info().Msg("Memcached connected!")
	return nil
}
