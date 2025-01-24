package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/codenotary/immudb/pkg/client"
)

// Payment é a estrutura do registro extraído do campo "after" do evento Kafka
type Payment struct {
	OrderNumber         string `json:"order_number"`
	PaymentAmount       string `json:"payment_amount"`
	TransactionAmount   string `json:"transaction_amount"`
	NameOnCard          string `json:"name_on_card"`
	CardNumber          string `json:"card_number"`
	ExpiryDate          string `json:"expiry_date"`
	SecurityCode        string `json:"security_code"`
	PostalCode          string `json:"postal_code"`
	TransactionDatetime int64  `json:"transaction_datetime"`
}

// KafkaEvent é a estrutura do evento Kafka recebido
type KafkaEvent struct {
	After Payment `json:"after"`
}

func main() {
	log.Println("Iniciando a aplicação Kafka -> ImmuDB")

	// Obter parâmetros do Kafka e do ImmuDB de variáveis de ambiente ou usar valores padrão
	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("KAFKA_TOPIC", "event.public.payments")
	kafkaGroup := getEnv("KAFKA_CONSUMER_GROUP", "payment-consumer-group")

	immuHost := getEnv("IMMUD_HOST", "localhost")
	immuPort := getEnvAsInt("IMMUD_PORT", 3322)
	immuUser := getEnv("IMMUD_USER", "immudb")
	immuPassword := getEnv("IMMUD_PASSWORD", "immudb")

	log.Printf("Configuração do Kafka - Brokers: %s, Tópico: %s, Grupo: %s", kafkaBrokers, kafkaTopic, kafkaGroup)
	log.Printf("Configuração do ImmuDB - Host: %s, Porta: %d", immuHost, immuPort)

	// Configuração do Kafka
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Version = sarama.V2_6_0_0

	log.Println("Inicializando o cliente ImmuDB...")
	consumer := &KafkaConsumer{
		immuClient: initializeImmuDB(immuHost, immuPort, immuUser, immuPassword), // Inicializa o cliente ImmuDB
	}

	for {
		log.Println("Conectando ao Kafka...")
		client, err := sarama.NewConsumerGroup([]string{kafkaBrokers}, kafkaGroup, config)
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
			if err := client.Consume(ctx, []string{kafkaTopic}, consumer); err != nil {
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
		client.Close()
	}
}

// KafkaConsumer representa o consumidor do Kafka
type KafkaConsumer struct {
	immuClient client.ImmuClient
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
		var event KafkaEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Erro ao decodificar mensagem: %v", err)
			continue
		}

		log.Printf("Mensagem decodificada com sucesso: %+v", event.After)

		// Insere o registro extraído no ImmuDB
		if err := kc.insertIntoImmuDB(event.After); err != nil {
			log.Printf("Erro ao inserir no ImmuDB: %v", err)
			continue
		}

		log.Printf("Registro inserido no ImmuDB com sucesso: OrderNumber=%s", event.After.OrderNumber)

		// Marca a mensagem como processada
		sess.MarkMessage(msg, "")
		log.Printf("Mensagem marcada como processada - Offset: %d", msg.Offset)
	}
	log.Printf("Finalizado o processamento de mensagens do tópico: %s", claim.Topic())
	return nil
}

// insertIntoImmuDB insere um registro no ImmuDB
func (kc *KafkaConsumer) insertIntoImmuDB(payment Payment) error {
	log.Printf("Preparando para inserir no ImmuDB: OrderNumber=%s", payment.OrderNumber)
	// Converte o registro em JSON
	data, err := json.Marshal(payment)
	if err != nil {
		return fmt.Errorf("erro ao converter registro para JSON: %w", err)
	}

	// Insere no ImmuDB como chave-valor
	_, err = kc.immuClient.Set(context.Background(), []byte(payment.OrderNumber), data)
	if err != nil {
		return fmt.Errorf("erro ao inserir no ImmuDB: %w", err)
	}

	log.Printf("Inserção no ImmuDB concluída: OrderNumber=%s", payment.OrderNumber)
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

// getEnv retorna o valor de uma variável de ambiente ou um valor padrão
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt retorna o valor de uma variável de ambiente como int ou um valor padrão
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
