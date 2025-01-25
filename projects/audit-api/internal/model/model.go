package model

import "time"

type AuditTrail struct {
	Application    string    `json:"application"`
	DbName         string    `json:"dbName"`
	DbSchema       string    `json:"dbSchema"`
	DbTable        string    `json:"dbTable"`
	EventOperation string    `json:"eventOperation"`
	EventDate      time.Time `json:"eventDate"`
	Event          string    `json:"event"`
}
