package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Waelson/audit/audit-api/internal/dao"
	"log"
	"net/http"
	"strings"
)

func NewAuditTrailHandler(d dao.AuditTrailDao) AuditTrailHandler {
	return &auditTrailHandler{dao: d}
}

type AuditTrailHandler interface {
	QueryAuditTrail() http.HandlerFunc
}

type auditTrailHandler struct {
	dao dao.AuditTrailDao
}

// QueryAuditTrail manipula as solicitações para consultar eventos de trilha de auditoria
func (a *auditTrailHandler) QueryAuditTrail() http.HandlerFunc {
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

		rows, err := a.dao.QueryAuditTrail(ctx, params)
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
