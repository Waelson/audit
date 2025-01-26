package db

import (
	"context"
	"fmt"
	"github.com/Waelson/audit/audit-api/pkg/config"
	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/codenotary/immudb/pkg/client"
	"log"
)

// NewImmuDBClient inicializa um novo cliente ImmuDB
func NewImmuDBClient(cfg config.ImmuDBConfig) (client.ImmuClient, error) {
	ctx := context.Background()
	log.Printf("Conectando ao ImmuDB em %s:%d...", cfg.Host, cfg.Port)

	immuClient, err := client.NewImmuClient(client.DefaultOptions().WithAddress(cfg.Host).WithPort(cfg.Port))
	if err != nil {
		fmt.Errorf("Falha ao conectar ao ImmuDB: %v", err)
		return nil, err
	}

	log.Println("Autenticando no ImmuDB...")
	_, err = immuClient.Login(ctx, []byte(cfg.User), []byte(cfg.Password))
	if err != nil {
		fmt.Errorf("Falha ao autenticar no ImmuDB: %v", err)
		return nil, err
	}

	// Usa o banco de dados criado
	_, err = immuClient.UseDatabase(context.Background(), &schema.Database{DatabaseName: cfg.Db})
	if err != nil {
		fmt.Errorf("erro ao usar banco de dados '%s': %w", cfg.Db, err)
		return nil, err
	}

	log.Println("Conexão e autenticação com o ImmuDB bem-sucedidas.")
	return immuClient, nil
}
