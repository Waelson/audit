package main

import (
	"context"
	"fmt"
	consumer2 "github.com/Waelson/audit/audit-consumer/internal/consumer"
	"github.com/Waelson/audit/audit-consumer/internal/utils"
	"github.com/codenotary/immudb/pkg/api/schema"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/codenotary/immudb/pkg/client"
)

func main() {
	log.Println("Iniciando a aplicação Kafka -> ImmuDB")

	// Obter parâmetros do Kafka e do ImmuDB de variáveis de ambiente ou usar valores padrão
	kafkaBrokers := utils.GetEnv("KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := utils.GetEnv("KAFKA_TOPIC", "audit-trail")
	kafkaGroup := utils.GetEnv("KAFKA_CONSUMER_GROUP", "audit-trail-consumer-group")

	immuHost := utils.GetEnv("IMMUD_HOST", "localhost")
	immuPort := utils.GetEnvAsInt("IMMUD_PORT", 3322)
	immuUser := utils.GetEnv("IMMUD_USER", "immudb")
	immuPassword := utils.GetEnv("IMMUD_PASSWORD", "immudb")

	log.Printf("Configuração do Kafka - Brokers: %s, Tópico: %s, Grupo: %s", kafkaBrokers, kafkaTopic, kafkaGroup)
	log.Printf("Configuração do ImmuDB - Host: %s, Porta: %d", immuHost, immuPort)

	// Inicializa o cliente ImmuDB
	immuClient := initializeImmuDB(immuHost, immuPort, immuUser, immuPassword)

	// Cria o banco de dados e tabela automaticamente
	if err := createDatabaseAndTable(immuClient, "audit_db"); err != nil {
		log.Fatalf("Erro ao configurar o banco de dados e tabela: %v", err)
	}

	log.Println("Inicializando o consumidor Kafka...")
	consumer := &consumer2.KafkaConsumer{
		ImmuClient: immuClient,
	}

	// Configuração do Kafka
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Version = sarama.V2_6_0_0

	for {
		log.Println("Conectando ao Kafka...")
		kafkaClient, err := sarama.NewConsumerGroup([]string{kafkaBrokers}, kafkaGroup, config)
		if err != nil {
			log.Printf("Erro ao criar consumer group: %v. Tentando novamente em 5 segundos...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Captura interrupções do sistema para fechar o consumidor com segurança
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			sigchan := make(chan os.Signal, 1)
			signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
			<-sigchan
			log.Println("Interrompendo o consumidor...")
			cancel()
		}()

		log.Printf("Consumindo mensagens do tópico: %s", kafkaTopic)
		for {
			if err := kafkaClient.Consume(ctx, []string{kafkaTopic}, consumer); err != nil {
				log.Printf("Erro ao consumir mensagens: %v. Tentando reconectar em 5 segundos...", err)
				time.Sleep(5 * time.Second)
				break
			}
			if ctx.Err() != nil {
				log.Println("Contexto encerrado, saindo do loop de consumo.")
				break
			}
		}
		log.Println("Fechando o cliente Kafka...")
		kafkaClient.Close()
	}
}

// createDatabaseAndTable cria o banco de dados e tabela se ainda não existirem
func createDatabaseAndTable(immuClient client.ImmuClient, dbName string) error {
	log.Printf("Criando banco de dados '%s' e tabela 'payments', se não existirem...", dbName)

	// Cria o banco de dados, se não existir
	_, err := immuClient.CreateDatabaseV2(context.Background(), dbName, nil)
	if err != nil {
		if err.Error() != "database already exists" {
			return fmt.Errorf("erro ao criar banco de dados: %w", err)
		}
	}

	// Usa o banco de dados criado
	_, err = immuClient.UseDatabase(context.Background(), &schema.Database{DatabaseName: dbName})
	if err != nil {
		return fmt.Errorf("erro ao usar banco de dados '%s': %w", dbName, err)
	}

	// Cria a tabela 'payments', se não existir
	query := `
		CREATE TABLE IF NOT EXISTS audit_trail (
			id INTEGER AUTO_INCREMENT, 
			connector VARCHAR,
			application VARCHAR,         
			db_name VARCHAR,           
			db_schema VARCHAR,         
			db_table VARCHAR,          
			event_operation VARCHAR,   
			event_date TIMESTAMP,      
			event JSON,                
			PRIMARY KEY (id)           
		);
	`
	_, err = immuClient.SQLExec(context.Background(), query, nil)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela no ImmuDB: %w", err)
	}

	log.Println("Banco de dados e tabela configurados com sucesso.")
	return nil
}

// initializeImmuDB inicializa o cliente ImmuDB
func initializeImmuDB(host string, port int, user, password string) client.ImmuClient {
	log.Printf("Inicializando conexão com o ImmuDB - Host: %s, Porta: %d", host, port)
	immuClient, err := client.NewImmuClient(client.DefaultOptions().WithAddress(host).WithPort(port))
	if err != nil {
		log.Fatalf("Erro ao conectar ao ImmuDB: %v", err)
	}

	log.Println("Autenticando no ImmuDB...")
	_, err = immuClient.Login(context.Background(), []byte(user), []byte(password))
	if err != nil {
		log.Fatalf("Erro ao autenticar no ImmuDB: %v", err)
	}

	log.Println("Conexão com o ImmuDB estabelecida com sucesso.")
	return immuClient
}
