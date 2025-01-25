package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/Waelson/audit/audit-consumer/internal/model"
	"github.com/codenotary/immudb/pkg/client"
	"log"
	"time"
)

// KafkaConsumer representa o consumidor do Kafka
type KafkaConsumer struct {
	ImmuClient client.ImmuClient
}

// Setup é executado antes de uma nova sessão de consumo
func (kc *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	log.Println("Setup da sessão do consumer iniciado.")
	return nil
}

// Cleanup é executado ao final da sessão de consumo
func (kc *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("Cleanup da sessão do consumer finalizado.")
	return nil
}

// ConsumeClaim processa as mensagens do tópico
func (kc *KafkaConsumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("Iniciando o processamento de mensagens do tópico: %s", claim.Topic())
	for msg := range claim.Messages() {
		log.Printf("Mensagem recebida - Partição: %d, Offset: %d, Valor: %s", msg.Partition, msg.Offset, string(msg.Value))

		// Decodifica o evento Kafka para a estrutura KafkaEvent
		var event model.KafkaEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Erro ao decodificar mensagem: %v", err)
			continue
		}

		log.Printf("Mensagem decodificada com sucesso: %+v", event.After)

		// Insere o registro extraído no ImmuDB
		if err := kc.insertIntoImmuDB(event); err != nil {
			log.Printf("Erro ao inserir no ImmuDB: %v", err)
			continue
		}

		log.Printf("Registro inserido no ImmuDB com sucesso")

		// Marca a mensagem como processada
		sess.MarkMessage(msg, "")
		log.Printf("Mensagem marcada como processada - Offset: %d", msg.Offset)
	}
	log.Printf("Finalizado o processamento de mensagens do tópico: %s", claim.Topic())
	return nil
}

// insertIntoImmuDB insere um evento Kafka na tabela audit_trail do ImmuDB
func (kc *KafkaConsumer) insertIntoImmuDB(event model.KafkaEvent) error {
	log.Printf("Preparando para inserir no ImmuDB: Operation=%s, Table=%s", event.Op, event.Source.Table)

	// Converter o timestamp do evento (TsMs) para time.Time
	eventDate := time.UnixMilli(event.TsMs)

	// Query para inserir dados na tabela audit_trail
	query := `
		INSERT INTO audit_trail (
			connector, application, db_name, db_schema, db_table, event_operation, event_date, event
		)
		VALUES (
			@connector, @application, @db_name, @db_schema, @db_table, @event_operation, @event_date, @event
		);
	`

	// Serializa o evento como string JSON
	e := model.Event{
		After:  event.After,
		Before: event.Before,
	}
	eventData, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("erro ao converter evento para JSON: %w", err)
	}

	// Cria o mapa de parâmetros para a query
	params := map[string]interface{}{
		"connector":       event.Source.Connector,
		"application":     event.Application,
		"db_name":         event.Source.Db,
		"db_schema":       event.Source.Schema,
		"db_table":        event.Source.Table,
		"event_operation": event.Op,
		"event_date":      eventDate,
		"event":           string(eventData),
	}

	// Executa a query SQL
	_, err = kc.ImmuClient.SQLExec(context.Background(), query, params)
	if err != nil {
		return fmt.Errorf("erro ao inserir evento no ImmuDB: %w", err)
	}

	log.Printf("Inserção no ImmuDB concluída: Table=%s, Operation=%s", event.Source.Table, event.Op)
	return nil
}
