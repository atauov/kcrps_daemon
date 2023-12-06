package service

import (
	"daemon"
	"daemon/pkg/repository"
)

type PosInvoice interface {
	SendInvoice(userId int, invoice daemon.Invoice) error
	CancelInvoice(userId, invoiceId int) error
	CancelPayment(userId, amount, isToday, invoiceId int) error
	CheckInvoices(userId, isToday int, invoices []string) error
	UpdateStatus(id, status, inWork int) error
	UpdateClientName(invoiceId int, clientName string) error
	GetInWorkInvoices(userId int) ([]daemon.Invoice, error)
	GetInvoiceAmount(invoiceId int) (int, error)
}

type Service struct {
	PosInvoice
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		PosInvoice: NewPosInvoiceService(repos.PosInvoice),
	}
}
