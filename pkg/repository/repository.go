package repository

import (
	"github.com/atauov/kcrps"
	"github.com/jmoiron/sqlx"
)

type Authorization interface {
	CreateUser(user kcrps.User) (int, error)
	GetUser(username, password string) (kcrps.User, error)
}

type Invoice interface {
	Create(userId int, invoice kcrps.Invoice) (int, error)
	GetAll(userId int) ([]kcrps.Invoice, error)
	GetById(userId, invoiceId int) (kcrps.Invoice, error)
	SetInvoiceForCancel(userId, invoiceId int) error
	SetInvoiceForRefund(userId, invoiceId int) error
}

type PosInvoice interface {
	UpdateStatus(id, status, inWork int) error
	UpdateClientName(invoiceId int, clientName string) error
	GetInWorkInvoices(userId int) ([]kcrps.Invoice, error)
	GetInvoiceAmount(invoiceId int) (int, error)
}

type Repository struct {
	Authorization
	Invoice
	PosInvoice
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
		Invoice:       NewInvoicePostgres(db),
		PosInvoice:    NewPosInvoicePostgres(db),
	}
}
