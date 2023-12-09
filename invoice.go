package daemon

import "time"

type Invoice struct {
	UUID       int       `json:"uuid" db:"uuid"`
	PosID      int       `json:"pos-id"`
	CreatedAt  time.Time `json:"created-at" db:"created_at"`
	Account    string    `json:"account" db:"account" binding:"required"`
	Amount     int       `json:"amount" db:"amount" binding:"required"`
	ClientName string    `json:"client-name" db:"client_name"`
	Message    string    `json:"message" db:"message"`
	Status     int       `json:"status" db:"status"`
}
