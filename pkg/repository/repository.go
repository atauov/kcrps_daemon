package repository

import (
	"daemon"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PosInvoice interface {
	UpdateStatus(invoice daemon.Invoice, status int) error
	UpdateClientName(invoice daemon.Invoice, clientName string) error
	GetInWorkInvoices(posTerminalId uuid.UUID) ([]daemon.Invoice, error)
	GetInvoiceAmount(invoice daemon.Invoice) (int, error)
	GetAllPosTerminals() ([]daemon.PosTerminal, error)
}

type Repository struct {
	PosInvoice
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		PosInvoice: NewPosInvoicePostgres(db),
	}
}
