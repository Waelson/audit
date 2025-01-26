package main

import (
	"github.com/Waelson/audit/audit-api/internal/dao"
	"github.com/Waelson/audit/audit-api/internal/handler"
	"github.com/Waelson/audit/audit-api/pkg/config"
	"github.com/Waelson/audit/audit-api/pkg/db"
	"github.com/Waelson/audit/audit-api/pkg/middleware"
	"log"
	"net/http"
)

func main() {
	log.Println("Iniciando servidor...")

	cfg := config.GetImmuDBConfig()
	log.Printf("Configuração do ImmuDB: %+v", cfg)

	dbClient, err := db.NewImmuDBClient(cfg)
	if err != nil {
		log.Fatalf("Falha ao criar o cliente ImmuDB: %v", err)
	}

	filterDao := dao.NewFilterDao(dbClient)
	auditTrailDao := dao.NewAuditTrailDao(dbClient)
	log.Println("DAOs iniciadas com sucesso.")

	filterHandler := handler.NewFilterHandler(filterDao)
	auditTrailHandler := handler.NewAuditTrailHandler(auditTrailDao)
	log.Println("Handlers iniciados com sucesso.")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/audit-trail", auditTrailHandler.QueryAuditTrail())
	mux.HandleFunc("/api/filters", filterHandler.QueryFilters())
	log.Println("Rotas registradas com sucesso.")

	// Adiciona o middleware de CORS
	hdl := middleware.CorsMiddleware(mux)

	log.Println("Servidor iniciado na porta :5050")
	if err := http.ListenAndServe(":5050", hdl); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
