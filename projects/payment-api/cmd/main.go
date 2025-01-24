package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/rs/cors" // Biblioteca de CORS
)

type Payment struct {
	OrderNumber         string  `json:"orderNumber"`
	PaymentAmount       float64 `json:"paymentAmount"`
	TransactionAmount   float64 `json:"transactionAmount"`
	NameOnCard          string  `json:"nameOnCard"`
	CardNumber          string  `json:"cardNumber"`
	ExpiryDate          string  `json:"expiryDate"`
	SecurityCode        string  `json:"securityCode"`
	PostalCode          string  `json:"postalCode"`
	TransactionDateTime string  `json:"transactionDateTime"`
}

var db *sql.DB

func main() {
	log.Println("Inicializando o servidor...")

	// Obtém os parâmetros do PostgreSQL das variáveis de ambiente ou usa valores padrão
	dbUser := getEnv("POSTGRES_USER", "postgres")
	dbPassword := getEnv("POSTGRES_PASSWORD", "password")
	dbName := getEnv("POSTGRES_DB", "payment_db")
	dbHost := getEnv("POSTGRES_HOST", "localhost")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	sslMode := getEnv("POSTGRES_SSLMODE", "disable")

	log.Printf("Configurando conexão com o banco de dados PostgreSQL (Host: %s, Port: %s, DB: %s)", dbHost, dbPort, dbName)

	// String de conexão ao PostgreSQL
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		dbUser, dbPassword, dbName, dbHost, dbPort, sslMode)

	log.Printf(connStr)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer func() {
		log.Println("Fechando conexão com o banco de dados...")
		db.Close()
	}()

	// Testa a conexão com o banco de dados
	log.Println("Verificando conexão com o banco de dados...")
	err = db.Ping()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	log.Println("Conexão com o banco de dados estabelecida com sucesso!")

	// Configuração do router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Configuração do middleware de CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://myfrontend.com"}, // Adicione as origens permitidas
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutos
	})

	// Aplica o middleware de CORS no router
	r.Use(corsMiddleware.Handler)

	// Rota para receber pagamentos
	r.Post("/api/v1/payment", handlePayment)

	// Inicializa o servidor na porta 8080
	log.Println("Servidor rodando na porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// handlePayment processa a requisição de pagamento e insere no banco
func handlePayment(w http.ResponseWriter, r *http.Request) {
	log.Println("Recebendo requisição de pagamento...")

	var payment Payment

	// Decodifica o corpo da requisição JSON
	log.Println("Decodificando dados da requisição...")
	err := json.NewDecoder(r.Body).Decode(&payment)
	if err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	// Valida e processa a data/hora da transação
	if payment.TransactionDateTime == "" {
		payment.TransactionDateTime = time.Now().Format("2006-01-02 15:04:05")
		log.Printf("Data/hora da transação não fornecida. Usando a hora atual: %s", payment.TransactionDateTime)
	}

	// Log dos dados recebidos
	log.Printf("Dados do pagamento recebidos: %+v", payment)

	// Insere os dados do pagamento no banco de dados
	log.Println("Inserindo dados no banco de dados...")
	query := `
		INSERT INTO payments (
			order_number, payment_amount, transaction_amount, name_on_card,
			card_number, expiry_date, security_code, postal_code, transaction_datetime
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = db.Exec(query,
		payment.OrderNumber,
		payment.PaymentAmount,
		payment.TransactionAmount,
		payment.NameOnCard,
		payment.CardNumber,
		payment.ExpiryDate,
		payment.SecurityCode,
		payment.PostalCode,
		payment.TransactionDateTime,
	)
	if err != nil {
		log.Printf("Erro ao salvar os dados no banco de dados: %v", err)
		http.Error(w, "Erro ao salvar os dados no banco de dados", http.StatusInternalServerError)
		return
	}

	log.Println("Dados inseridos com sucesso no banco de dados.")

	// Retorna uma resposta de sucesso
	log.Println("Enviando resposta de sucesso ao cliente...")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Pagamento recebido com sucesso!"))
	log.Println("Requisição de pagamento processada com sucesso.")
}

// getEnv busca o valor de uma variável de ambiente ou retorna o padrão
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("Variável de ambiente %s não configurada. Usando valor padrão: %s", key, defaultValue)
	return defaultValue
}
