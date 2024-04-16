package service

import (
	"daemon"
	"daemon/pkg/repository"
)

type PosInvoice interface {
	SendInvoice(invoice daemon.Invoice, posTerminal daemon.PosTerminal) error
	CancelInvoice(invoice daemon.Invoice, posTerminal daemon.PosTerminal) error
	//CancelPayment(invoice daemon.Invoice, posTerminal daemon.PosTerminal, amount, isToday int) error
	CheckInvoices(posTerminal daemon.PosTerminal, isToday int, invoices []string) error
	UpdateStatus(invoice daemon.Invoice, status int) error
	UpdateClientName(invoice daemon.Invoice, clientName string) error
	GetInWorkInvoices(posTerminal daemon.PosTerminal) ([]daemon.Invoice, error)
	GetInvoiceAmount(invoice daemon.Invoice) (int, error)
	GetAllPosTerminals() ([]daemon.PosTerminal, error)
	SetOldInvoicesToCancel() error
}

type Service struct {
	PosInvoice
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		PosInvoice: NewPosInvoiceService(repos.PosInvoice),
	}
}
