package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config sdk for getting key values from settings file
type Config struct {
	viper *viper.Viper
}

// NewConfig initializes configuration from settings file
func NewConfig(defaults map[string]string) *Config {
	config := &Config{viper: viper.New()}

	config.viper.SetDefault("JWT_SECRET", "Secret key for test")

	config.viper.SetEnvPrefix("ISLA")
	config.viper.AutomaticEnv()

	config.viper.SetConfigName("settings")
	config.viper.AddConfigPath(".")

	err := config.viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Can not load config file: %s", err))
	}

	return config
}

//IsSet checks if a key passed as a parameter has been set a value
func (config *Config) IsSet(key string) bool {
	return config.viper.IsSet(key)
}

//GetBool returns boolean value set for a key passed to this function
func (config *Config) GetBool(key string) bool {
	return config.viper.GetBool(key)
}

//GetString return string value set for a key passed to this function
func (config *Config) GetString(key string) string {
	return config.viper.GetString(key)
}

//GetInt return int value set for a key passed to this function
func (config *Config) GetInt(key string) int {
	return config.viper.GetInt(key)
}

//GetMapString returns the value associated with the key as a map of strings
func (config *Config) GetMapString(key string) map[string]string {
	return config.viper.GetStringMapString(key)
}

//GetMap returns the value associated with the key as a map of interfaces
func (config *Config) GetMap(key string) map[string]interface{} {
	return config.viper.GetStringMap(key)
}
