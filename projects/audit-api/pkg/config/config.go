package config

import (
	"github.com/Waelson/audit/audit-api/pkg/utils"
	"log"
)

// Configuração padrão para conexão ao ImmuDB
const (
	defaultHost     = "localhost"
	defaultPort     = 3322
	defaultUser     = "immudb"
	defaultPassword = "immudb"
	defaultDb       = "audit_db"
)

// Configuração da conexão com o ImmuDB
type ImmuDBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Db       string
}

func GetImmuDBConfig() ImmuDBConfig {
	log.Println("Obtendo configuração do ImmuDB a partir das variáveis de ambiente...")
	return ImmuDBConfig{
		Host:     utils.GetEnv("IMMUD_HOST", defaultHost),
		Port:     utils.GetEnvAsInt("IMMUD_PORT", defaultPort),
		User:     utils.GetEnv("IMMUD_USER", defaultUser),
		Password: utils.GetEnv("IMMUD_PASSWORD", defaultPassword),
		Db:       utils.GetEnv("IMMUD_DB", defaultDb),
	}
}
