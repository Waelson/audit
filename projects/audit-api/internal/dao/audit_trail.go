package dao

import (
	"context"
	"fmt"
	"github.com/Waelson/audit/audit-api/internal/model"
	"github.com/codenotary/immudb/pkg/client"
	"log"
	"time"
)

type AuditTrailDao interface {
	QueryAuditTrail(ctx context.Context, params map[string]interface{}) ([]model.AuditTrail, error)
}

type auditTrailDao struct {
	client client.ImmuClient
}

func NewAuditTrailDao(client client.ImmuClient) AuditTrailDao {
	return &auditTrailDao{client: client}
}

func (db *auditTrailDao) QueryAuditTrail(ctx context.Context, params map[string]interface{}) ([]model.AuditTrail, error) {
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
