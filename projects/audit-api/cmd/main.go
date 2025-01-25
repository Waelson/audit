package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Waelson/audit/audit-api/internal/model"
	"github.com/Waelson/audit/audit-api/internal/utils"
	"github.com/codenotary/immudb/pkg/api/schema"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/codenotary/immudb/pkg/client"
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

func getImmuDBConfig() ImmuDBConfig {
	log.Println("Obtendo configuração do ImmuDB a partir das variáveis de ambiente...")
	return ImmuDBConfig{
		Host:     utils.GetEnv("IMMUD_HOST", defaultHost),
		Port:     utils.GetEnvAsInt("IMMUD_PORT", defaultPort),
		User:     utils.GetEnv("IMMUD_USER", defaultUser),
		Password: utils.GetEnv("IMMUD_PASSWORD", defaultPassword),
		Db:       utils.GetEnv("IMMUD_DB", defaultDb),
	}
}

// ImmuDBClient encapsula a lógica de acesso ao ImmuDB
type ImmuDBClient struct {
	client client.ImmuClient
}

// NewImmuDBClient inicializa um novo cliente ImmuDB
func NewImmuDBClient(config ImmuDBConfig) *ImmuDBClient {
	ctx := context.Background()
	log.Printf("Conectando ao ImmuDB em %s:%d...", config.Host, config.Port)

	immuClient, err := client.NewImmuClient(client.DefaultOptions().WithAddress(config.Host).WithPort(config.Port))
	if err != nil {
		log.Fatalf("Falha ao conectar ao ImmuDB: %v", err)
	}

	log.Println("Autenticando no ImmuDB...")
	_, err = immuClient.Login(ctx, []byte(config.User), []byte(config.Password))
	if err != nil {
		log.Fatalf("Falha ao autenticar no ImmuDB: %v", err)
	}

	// Usa o banco de dados criado
	_, err = immuClient.UseDatabase(context.Background(), &schema.Database{DatabaseName: config.Db})
	if err != nil {
		fmt.Errorf("erro ao usar banco de dados '%s': %w", config.Db, err)
	}

	log.Println("Conexão e autenticação com o ImmuDB bem-sucedidas.")
	return &ImmuDBClient{client: immuClient}
}

// QueryAuditTrail executa a consulta de eventos na tabela audit_trail
func (db *ImmuDBClient) QueryAuditTrail(ctx context.Context, params map[string]interface{}) ([]model.AuditTrail, error) {
	log.Printf("Executando consulta de audit trail com parâmetros: %+v", params)
	query := `
		SELECT application, db_name, db_schema, db_table, event_operation, event_date, event 
		FROM audit_trail 
		SINCE @start_date UNTIL @end_date
		WHERE application = @application AND db_name = @db_name AND db_schema = @db_schema AND db_table = @db_table AND event_operation = @event_operation;
	`

	sqlResult, err := db.client.SQLQuery(ctx, query, params, false)
	if err != nil {
		log.Printf("Erro ao executar a consulta de audit trail: %v", err)
		return nil, fmt.Errorf("error querying audit trail: %w", err)
	}

	log.Println("Consulta de audit trail executada com sucesso.")
	response := make([]model.AuditTrail, 0)
	for _, row := range sqlResult.Rows {
		//event := strings.ReplaceAll(row.Values[6].GetS(), "\\", "")
		//event = event[0:]
		trail := model.AuditTrail{
			Application:    row.Values[0].GetS(),
			DbName:         row.Values[1].GetS(),
			DbSchema:       row.Values[2].GetS(),
			DbTable:        row.Values[3].GetS(),
			EventOperation: row.Values[4].GetS(),
			EventDate:      time.UnixMicro(row.Values[5].GetTs()),
			Event:          row.Values[6].GetS(),
		}
		response = append(response, trail)
	}

	return response, nil
}

// QueryFilters executa a consulta para obter filtros hierárquicos e retorna no formato especificado
func (db *ImmuDBClient) QueryFilters(ctx context.Context) ([]map[string]interface{}, error) {
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

// Handlers

// queryAuditTrailHandler manipula as solicitações para consultar eventos de trilha de auditoria
func queryAuditTrailHandler(db *ImmuDBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Recebendo solicitação para consultar audit trail...")
		ctx := context.Background()
		application := r.URL.Query().Get("application")
		dbName := r.URL.Query().Get("db_name")
		dbSchema := r.URL.Query().Get("db_schema")
		dbTable := r.URL.Query().Get("db_table")
		startDate := r.URL.Query().Get("start_date")
		endDate := r.URL.Query().Get("end_date")
		eventOperation := strings.ToLower(r.URL.Query().Get("event_operation"))

		startDate = strings.ReplaceAll(startDate, "T", " ") + ":00"
		endDate = strings.ReplaceAll(endDate, "T", " ") + ":00"

		if application == "" || dbName == "" || dbSchema == "" || dbTable == "" || startDate == "" || endDate == "" || eventOperation == "" {
			log.Println("Parâmetros de consulta ausentes na solicitação.")
			http.Error(w, "Missing required query parameters", http.StatusBadRequest)
			return
		}

		params := map[string]interface{}{
			"application":     application,
			"db_name":         dbName,
			"db_schema":       dbSchema,
			"db_table":        dbTable,
			"start_date":      startDate,
			"end_date":        endDate,
			"event_operation": eventOperation,
		}

		rows, err := db.QueryAuditTrail(ctx, params)
		if err != nil {
			log.Printf("Erro ao consultar audit trail: %v", err)
			http.Error(w, fmt.Sprintf("Error querying audit trail: %v", err), http.StatusInternalServerError)
			return
		}

		log.Println("Consulta de audit trail bem-sucedida, enviando resposta.")
		jsonResult, err := json.Marshal(rows)
		if err != nil {
			log.Printf("Erro ao serializar a resposta JSON: %v", err)
			http.Error(w, fmt.Sprintf("Error encoding result to JSON: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResult)
	}
}

// queryFiltersHandler manipula as solicitações para obter filtros hierárquicos
func queryFiltersHandler(db *ImmuDBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Recebendo solicitação para obter filtros hierárquicos...")
		ctx := context.Background()
		filters, err := db.QueryFilters(ctx)
		if err != nil {
			log.Printf("Erro ao consultar filtros: %v", err)
			http.Error(w, fmt.Sprintf("Error querying filters: %v", err), http.StatusInternalServerError)
			return
		}

		log.Println("Consulta de filtros bem-sucedida, enviando resposta.")
		jsonResult, err := json.Marshal(filters)
		if err != nil {
			log.Printf("Erro ao serializar a resposta JSON: %v", err)
			http.Error(w, fmt.Sprintf("Error encoding filters to JSON: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResult)
	}
}

// Middleware para adicionar cabeçalhos de CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	log.Println("Iniciando servidor...")

	config := getImmuDBConfig()
	log.Printf("Configuração do ImmuDB: %+v", config)

	dbClient := NewImmuDBClient(config)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/audit-trail", queryAuditTrailHandler(dbClient))
	mux.HandleFunc("/api/filters", queryFiltersHandler(dbClient))

	// Adiciona o middleware de CORS
	handler := corsMiddleware(mux)

	log.Println("Servidor iniciado na porta :5050")
	if err := http.ListenAndServe(":5050", handler); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
