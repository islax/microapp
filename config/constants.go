package config

const (
	// EvPrefix environment variable prefix
	EvPrefix = "ISLA"

	// EvSuffixForAPIClientHTTPTimeout environment variable name for API client http timeout
	EvSuffixForAPIClientHTTPTimeout = "APICLIENT_HTTP_TIMEOUT"
	// EvSuffixForDBHost environment variable name for database host
	EvSuffixForDBHost = "DB_HOST"
	// EvSuffixForDBConnectionLifetime environment variable name for connection lifetime in database connection pool
	EvSuffixForDBConnectionLifetime = "DB_CONNECTION_MAX_LIFETIME"
	// EvSuffixForDBLogLevel environment variable name for database log level
	EvSuffixForDBLogLevel = "DB_LOG_LEVEL"
	// EvSuffixForDBMaxIdleConnections environment variable name for max idle connections in database connection pool
	EvSuffixForDBMaxIdleConnections = "DB_MAX_IDLE_CONNECTIONS"
	// EvSuffixForDBMaxOpenConnections environment variable name for max open connections in database connection pool
	EvSuffixForDBMaxOpenConnections = "DB_MAX_OPEN_CONNECTIONS"
	// EvSuffixForDBPassword environment variable name for database bind user password
	EvSuffixForDBPassword = "DB_PWD"
	// EvSuffixForDBPort environment variable name for database port
	EvSuffixForDBPort = "DB_PORT"
	// EvSuffixForDBRequired environment variable name for database required flag
	EvSuffixForDBRequired = "DB_REQUIRED"
	// EvSuffixForDBUser environment variable name for database bind user
	EvSuffixForDBUser = "DB_USER"
	// EvSuffixForHTTPIdleTimeout environment variable name for HTT idle timeout
	EvSuffixForHTTPIdleTimeout = "HTTP_IDLE_TIMEOUT"
	// EvSuffixForHTTPReadTimeout environment variable name for HTTP read timeout
	EvSuffixForHTTPReadTimeout = "HTTP_READ_TIMEOUT"
	// EvSuffixForHTTPWriteTimeout environment variable name for http write timeout
	EvSuffixForHTTPWriteTimeout = "HTTP_WRITE_TIMEOUT"
	// EvSuffixForJwtSecret environment variable name for JWT secrete
	EvSuffixForJwtSecret = "JWT_SECRET"
	// EvSuffixForLogLevel environment variable name for log level
	EvSuffixForLogLevel = "LOG_LEVEL"
	// EvSuffixForGormSlowThreshold environment variable name for log level
	EvSuffixForGormSlowThreshold = "GORM_SLOW_THRESHOLD"
	// EvSuffixForEnableHealthLog environment variable name for log level
	EvSuffixForEnableHealthLog = "ENABLE_HEALTH_LOG"
	// EvSuffixForEnableMetrics environment variable name for enable metrics
	EvSuffixForEnableMetrics = "ENABLE_METRICS"
	// EvSuffixForGormMetricsRefresh environment variable name for gorm metrics refresh interval
	EvSuffixForGormMetricsRefresh = "GORM_METRICS_REFRESH_INTERVAL"
)
