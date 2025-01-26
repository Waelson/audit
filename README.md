# Solução de Auditoria Baseada em CDC

![Golang](https://img.shields.io/badge/technology-Golang-blue.svg)  ![Debezium](https://img.shields.io/badge/technology-Debezium-orange.svg) ![React](https://img.shields.io/badge/technology-React-green.svg)

O texto vem aqui

## Arquitetura

![Architecture](documentation/architecture.png)

### Componentes


### Tecnnologias

## Instalação
A aplicação está configurada para ser executada com Docker Compose. Siga os passos logo abaixo, mas assegure-se de ter os pré-requisitos instalados:

**Pré-requisitos:**
- Go 1.21 ou superior
- Node 20.12 ou superior (para executar projeto localmente)

1. **Clona o repositório**

```bash
git clone https://github.com/Waelson/audit.git
cd audit
```

2. **Inicializa a stack**

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

4. **Acessa a aplicação**

## Interface de Usuário

### Simulador de Pagamentos

![payment](documentation/payment_simulator.png)

### Consulta de Trilhas de Auditoria

![audit](documentation/audit_trail_query.png)

## 🧠 Teoria

### 💡CDC - Change Data Capture

___

#### O que é?

#### Quais problemas ela ajuda a resolver?