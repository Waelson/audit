package utils

import (
	"os"
	"strconv"
)

// GetEnv retorna o valor de uma variável de ambiente ou um valor padrão
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// GetEnvAsInt retorna o valor de uma variável de ambiente como int ou um valor padrão
func GetEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
