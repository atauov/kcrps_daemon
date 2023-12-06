package service

import (
	"daemon"
	"daemon/pkg/repository"
	"github.com/google/uuid"
)

type PosInvoice interface {
	SendInvoice(posTerminal daemon.PosTerminal, invoice daemon.Invoice) error
	CancelInvoice(posTerminal daemon.PosTerminal, invoiceId int) error
	CancelPayment(posTerminal daemon.PosTerminal, amount, isToday, invoiceId int) error
	CheckInvoices(posTerminal daemon.PosTerminal, isToday int, invoices []string) error
	UpdateStatus(id, status, inWork int) error
	UpdateClientName(invoiceId int, clientName string) error
	GetInWorkInvoices(posTerminalId uuid.UUID) ([]daemon.Invoice, error)
	GetInvoiceAmount(invoiceId int) (int, error)
	GetAllPosTerminals() ([]daemon.PosTerminal, error)
}

type Service struct {
	PosInvoice
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		PosInvoice: NewPosInvoiceService(repos.PosInvoice),
	}
}
