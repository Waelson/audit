# Solução de Auditoria Baseada em CDC

![Golang](https://img.shields.io/badge/technology-Golang-blue.svg)  ![Debezium](https://img.shields.io/badge/technology-Debezium-orange.svg) ![React](https://img.shields.io/badge/technology-React-green.svg)

Este repositório contém a implementação de um sistema de auditoria de dados utilizando a técnica de Change Data Capture (CDC). O objetivo é capturar e registrar alterações realizadas em bancos de dados relacionais de forma transparente para a aplicação, oferecendo uma solução prática para auditoria e rastreamento de eventos.

### Motivação

Sistemas onde a rastreabilidade de dados é essencial, como ambientes financeiros, e-commerce, ou qualquer aplicação sujeita a regulamentações, a auditoria de alterações nos dados é uma necessidade crítica. Este projeto demonstra como o CDC pode ser integrado a uma arquitetura moderna para capturar, armazenar e consultar eventos de forma estruturada.


## Arquitetura

![Architecture](documentation/architecture.png)

### Componentes

| **Componente**         | **Descrição**                                                                                                                              |
|------------------------|--------------------------------------------------------------------------------------------------------------------------------------------|
| **Payment UI**         | Interface gráfica (frontend) que permite ao usuário iniciar pagamentos.                                                                    |
| **Payment API**        | Serviço de backend responsável por processar e registrar transações no banco de dados PostgreSQL.                                          |
| **PostgreSQL**         | Banco de dados utilizado para armazenar os registros das transações.                                                                       |
| **Debezium Connector** | Plataforma de Change Data Capture (CDC) que monitora o PostgreSQL por meio de seus arquivos WAL (Write-Ahead Logs) para capturar mudanças. |
| **Kafka Cluster**      | Sistema de mensageria que recebe as alterações capturadas pelo Debezium e distribui mensagens para os consumidores.                        |
| **Audit Consumer**     | Serviço que consome as mensagens do Kafka, realiza transformações nos dados e os insere no banco de dados ImmuDB.                          |
| **ImmuDB**             | Banco de dados imutável utilizado para armazenar os registros de auditoria, garantindo integridade e rastreabilidade dos eventos.          |
| **Audit API**          | Serviço de backend que expõe endpoints para consulta e acesso aos registros de auditoria armazenados no ImmuDB.                            |
| **Audit UI**           | Interface gráfica (frontend) que permite a consulta e análise das trilhas de auditoria.                                                    |


### Tecnologias
| **Categoria**       | **Ferramenta/Descrição**    |
|---------------------|-----------------------------|
| **Linguagem**       | Golang, JavaScript (NodeJS) |
| **Bibliotecas**     | React                       |
| **Banco de Dados**  | ImmuDB, Postgres            |
| **Mensageria**      | Apache Kafka                |

## Instalação
A aplicação está configurada para ser executada com Docker Compose. Siga os passos logo abaixo, mas assegure-se de ter os pré-requisitos instalados:

**Pré-requisitos:**
- Docker
- Go 1.21 ou superior
- Node 20.12 ou superior (para executar projeto localmente)

1. **Clonar o repositório**

```bash
git clone https://github.com/Waelson/audit.git
cd audit
```

2. **Inicializar a stack**

```bash
docker-compose up --build
```

3. **Criar o conector**

```bash
curl --location 'http://localhost:8083/connectors' \
--header 'Content-Type: application/json' \
--data '{
  "name": "postgres-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "postgres",
    "database.port": "5432",
    "database.user": "postgres",
    "database.password": "password",
    "database.dbname": "payment_db",
    "database.server.name": "payment-postgres",
    "slot.name": "debezium_slot",
    "plugin.name": "pgoutput",
    "publication.name": "audit_events",
    "database.history.kafka.bootstrap.servers": "kafka:9092",
    "database.history.kafka.topic": "schema-changes.audit-trail",
    "topic.prefix": "audit",
    "table.include.list": "public.payments",
    "transforms": "RouteToTopic,AddAppName",
    "transforms.RouteToTopic.type": "org.apache.kafka.connect.transforms.RegexRouter",
    "transforms.RouteToTopic.regex": "audit.public.payments",
    "transforms.RouteToTopic.replacement": "audit-trail",
    "decimal.handling.mode": "string",
    "transforms.AddAppName.type": "org.apache.kafka.connect.transforms.InsertField$Value",
    "transforms.AddAppName.static.field": "application",
    "transforms.AddAppName.static.value": "payment-api"
  }
}
```

4. **Acessar o Simulador de Pagamento**
- Digite a URL http://localhost:3000/ no browser.
- Realize simulações de transações de pagamento de cartão de crédito clicando no botão `Pay`.

5. **Acessar a UI de Consulta de Trilhas de Auditoria**
- Digite a URL http://localhost:4000/ no browser.
- Preencha os filtros, lembrando de que a única operação contemplada é `Create` e deploy clique no botão `Search`. 

## Interface de Usuário

### Simulador de Pagamentos

![payment](documentation/payment_simulator.png)

### Consulta de Trilhas de Auditoria

![audit](documentation/audit_trail_query.png)

## Contribuições

Contribuições são bem-vindas! Sinta-se à vontade para abrir issues ou enviar pull requests com melhorias, correções ou novas funcionalidades.

## Licença

Este projeto está licenciado sob a Licença MIT.