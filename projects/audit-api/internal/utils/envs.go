package utils

import (
	"log"
	"os"
	"strconv"
)

// GetEnv Função utilitária para obter variáveis de ambiente com valor padrão
func GetEnv(key, defaultValue string) string {
	log.Printf("Obtendo variável de ambiente %s...", key)
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("Variável %s não encontrada, usando valor padrão: %s", key, defaultValue)
	return defaultValue
}

// GetEnvAsInt Função utilitária para obter variáveis de ambiente como inteiro com valor padrão
func GetEnvAsInt(key string, defaultValue int) int {
	log.Printf("Obtendo variável de ambiente %s como inteiro...", key)
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
