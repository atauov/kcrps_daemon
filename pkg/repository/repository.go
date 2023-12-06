package repository

import (
	"daemon"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PosInvoice interface {
	UpdateStatus(id, status, inWork int) error
	UpdateClientName(invoiceId int, clientName string) error
	GetInWorkInvoices(posTerminalId uuid.UUID) ([]daemon.Invoice, error)
	GetInvoiceAmount(invoiceId int) (int, error)
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
