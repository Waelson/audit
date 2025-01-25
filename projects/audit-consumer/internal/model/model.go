package model

type Source struct {
	Version   string `json:"version"`
	Connector string `json:"connector"`
	Name      string `json:"name"`
	TsMs      int64  `json:"ts_ms"`
	Snapshot  string `json:"snapshot"`
	Db        string `json:"db"`
	Sequence  string `json:"sequence"`
	TsUs      int64  `json:"ts_us"`
	TsNs      int64  `json:"ts_ns"`
	Schema    string `json:"schema"`
	Table     string `json:"table"`
	TxID      int    `json:"txId"`
	Lsn       int    `json:"lsn"`
}

// KafkaEvent é a estrutura do evento Kafka recebido
type KafkaEvent struct {
	Op          string      `json:"op"`
	TsMs        int64       `json:"ts_ms"`
	TsUs        int64       `json:"ts_us"`
	TsNs        int64       `json:"ts_ns"`
	After       interface{} `json:"after"`
	Before      interface{} `json:"before"`
	Source      Source      `json:"source"`
	Application string      `json:"application"`
}

type Event struct {
	After  interface{} `json:"after"`
	Before interface{} `json:"before"`
}
