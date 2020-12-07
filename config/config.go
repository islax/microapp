package config

import (
	"github.com/spf13/viper"
)

// Config sdk for getting key values from settings file
type Config struct {
	viper *viper.Viper
}

// NewConfig initializes configuration from settings file
func NewConfig(defaults map[string]interface{}) *Config {
	config := &Config{viper: viper.New()}

	config.viper.SetDefault(EvSuffixForJwtSecret, "Secret key for test")

	config.viper.SetDefault(EvSuffixForDBRequired, true)
	config.viper.SetDefault(EvSuffixForDBHost, "localhost")
	config.viper.SetDefault(EvSuffixForDBPort, "3306")
	config.viper.SetDefault(EvSuffixForDBUser, "root")
	config.viper.SetDefault(EvSuffixForDBPassword, "Cyber!nc#")

	config.viper.SetDefault(EvSuffixForLogLevel, "error")

	config.viper.SetDefault(EvSuffixForHTTPWriteTimeout, 15)
	config.viper.SetDefault(EvSuffixForHTTPReadTimeout, 15)
	config.viper.SetDefault(EvSuffixForHTTPIdleTimeout, 60)

	config.viper.SetDefault(EvSuffixForHTTPIdleTimeout, 15)

	for key, value := range defaults {
		config.viper.SetDefault(key, value)
	}

	config.viper.SetEnvPrefix(EvPrefix)
	config.viper.AutomaticEnv()

	// config.viper.SetConfigName("settings")
	// config.viper.AddConfigPath(".")

	// err := config.viper.ReadInConfig()
	// if err != nil {
	// 	fmt.Printf("Can not load config file: %s", err)
	// }

	return config
}

// IsSet checks if the give key's value been set
func (config *Config) IsSet(key string) bool {
	return config.viper.IsSet(key)
}

// GetBool returns boolean value set for the given key
func (config *Config) GetBool(key string) bool {
	return config.viper.GetBool(key)
}

// GetBoolWithDefault returns boolean value set for the given key, if not set returns the given defaultVal
func (config *Config) GetBoolWithDefault(key string, defaultVal bool) bool {
	if config.viper.IsSet(key) {
		return config.viper.GetBool(key)
	}
	return defaultVal
}

// GetString return string value set for a given key
func (config *Config) GetString(key string) string {
	return config.viper.GetString(key)
}

// GetStringWithDefault return string value set for the given key, if not set returns the given defaultVal
func (config *Config) GetStringWithDefault(key string, defaultVal string) string {
	if config.viper.IsSet(key) {
		return config.viper.GetString(key)
	}
	return defaultVal
}

// GetInt return int value set for the given key
func (config *Config) GetInt(key string) int {
	return config.viper.GetInt(key)
}

// GetMapString returns the value associated with the given key as a map of strings
func (config *Config) GetMapString(key string) map[string]string {
	return config.viper.GetStringMapString(key)
}

// GetMap returns the value associated with the given key as a map of interfaces
func (config *Config) GetMap(key string) map[string]interface{} {
	return config.viper.GetStringMap(key)
}

// Set sets the value for the given key
func (config *Config) Set(key string, value interface{}) {
	config.viper.Set(key, value)
}
