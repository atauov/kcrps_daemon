package repository

import (
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/atauov/kcrps"
	"github.com/jmoiron/sqlx"
)

type PosInvoicePostgres struct {
	db *sqlx.DB
}

func NewPosInvoicePostgres(db *sqlx.DB) *PosInvoicePostgres {
	return &PosInvoicePostgres{db: db}
}

func (r *PosInvoicePostgres) UpdateStatus(id, status, inWork int) error {
	query := fmt.Sprintf(`UPDATE %s SET status=$1, in_work=$2 WHERE id = $3`, invoicesTable)
	_, err := r.db.Exec(query, status, inWork, id)
	logrus.Printf("NEW STATUS = %d, IN_WORK = %d", status, inWork)
	return err
}

func (r *PosInvoicePostgres) UpdateClientName(invoiceId int, clientName string) error {
	query := fmt.Sprintf(`UPDATE %s SET client_name=$1 WHERE id = $2`, invoicesTable)
	_, err := r.db.Exec(query, clientName, invoiceId)
	return err
}

func (r *PosInvoicePostgres) GetInWorkInvoices(userId int) ([]kcrps.Invoice, error) {
	var invoices []kcrps.Invoice

	query := fmt.Sprintf("SELECT il.id, il.uuid, il.status, il.amount, il.account, il.message, il.in_work FROM %s il "+
		"INNER JOIN %s ul on il.id=ul.invoice_id WHERE ul.user_id = $1 AND il.in_work=1 ORDER BY il.id",
		invoicesTable, usersInvoicesTable)
	err := r.db.Select(&invoices, query, userId)

	return invoices, err
}

func (r *PosInvoicePostgres) GetInvoiceAmount(invoiceId int) (int, error) {
	var amount int
	query := fmt.Sprintf(`SELECT amount FROM %s WHERE id=$1`, invoicesTable)
	if err := r.db.Get(&amount, query, invoiceId); err != nil {
		return 0, err
	}
	return amount, nil
}
