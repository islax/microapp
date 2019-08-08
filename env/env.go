package env

import (
	"os"
)

//GetEnv returns the value of the given key, if no value found it returns the given default value
func GetEnv(key string, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultValue
}
