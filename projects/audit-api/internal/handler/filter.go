package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Waelson/audit/audit-api/internal/dao"
	"log"
	"net/http"
)

func NewFilterHandler(dao dao.FilterDao) FilterHandler {
	return &filterHandler{dao: dao}
}

type FilterHandler interface {
	QueryFilters() http.HandlerFunc
}

type filterHandler struct {
	dao dao.FilterDao
}

// QueryFilters manipula as solicitações para obter filtros hierárquicos
func (h *filterHandler) QueryFilters() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Recebendo solicitação para obter filtros hierárquicos...")
		ctx := context.Background()
		filters, err := h.dao.GetAll(ctx)
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
