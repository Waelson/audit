package dao

import (
	"context"
	"fmt"
	"github.com/codenotary/immudb/pkg/client"
	"log"
)

func NewFilterDao(client client.ImmuClient) FilterDao {
	return &filterDao{client: client}
}

type FilterDao interface {
	GetAll(ctx context.Context) ([]map[string]interface{}, error)
}

type filterDao struct {
	client client.ImmuClient
}

// GetAll executa a consulta para obter filtros hierárquicos e retorna no formato especificado
func (db *filterDao) GetAll(ctx context.Context) ([]map[string]interface{}, error) {
	log.Println("Executando consulta para obter filtros hierárquicos...")
	query := `
		SELECT application, db_name, db_schema, db_table 
		FROM audit_trail 
		GROUP BY application, db_name, db_schema, db_table 
		ORDER BY application, db_name, db_schema, db_table;
	`

	sqlResult, err := db.client.SQLQuery(ctx, query, nil, false)
	if err != nil {
		log.Printf("Erro ao executar a consulta de filtros: %v", err)
		return nil, fmt.Errorf("error querying filters: %w", err)
	}

	log.Println("Consulta de filtros executada com sucesso.")

	// Estrutura para armazenar o resultado
	filterMap := make(map[string]map[string]map[string][]string)

	// Processa as linhas da consulta
	for _, row := range sqlResult.Rows {
		app := row.Values[0].GetS()
		db := row.Values[1].GetS()
		schema := row.Values[2].GetS()
		table := row.Values[3].GetS()

		if _, exists := filterMap[app]; !exists {
			filterMap[app] = make(map[string]map[string][]string)
		}
		if _, exists := filterMap[app][db]; !exists {
			filterMap[app][db] = make(map[string][]string)
		}
		filterMap[app][db][schema] = append(filterMap[app][db][schema], table)
	}

	// Converte o resultado para o formato especificado
	var result []map[string]interface{}
	for app, dbMap := range filterMap {
		appEntry := map[string]interface{}{
			"application": app,
			"databases":   []map[string]interface{}{},
		}

		for dbName, schemaMap := range dbMap {
			dbEntry := map[string]interface{}{
				"name":    dbName,
				"schemas": []map[string]interface{}{},
			}

			for schemaName, tables := range schemaMap {
				schemaEntry := map[string]interface{}{
					"name":   schemaName,
					"tables": tables,
				}
				dbEntry["schemas"] = append(dbEntry["schemas"].([]map[string]interface{}), schemaEntry)
			}

			appEntry["databases"] = append(appEntry["databases"].([]map[string]interface{}), dbEntry)
		}

		result = append(result, appEntry)
	}

	return result, nil
}
