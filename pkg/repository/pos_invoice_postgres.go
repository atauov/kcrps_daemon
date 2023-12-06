package repository

import (
	"daemon"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
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

func (r *PosInvoicePostgres) GetInWorkInvoices(posTerminalId uuid.UUID) ([]daemon.Invoice, error) {
	var invoices []daemon.Invoice

	query := fmt.Sprintf("SELECT id, uuid, status, amount, account, message FROM %s il WHERE pos_id = $1 "+
		"AND status IN(1, 2, 4, 6, 10) ORDER BY created_at", invoicesTable)
	err := r.db.Select(&invoices, query, posTerminalId)

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

func (r *PosInvoicePostgres) GetAllPosTerminals() ([]daemon.PosTerminal, error) {
	var terminals []daemon.PosTerminal
	query := fmt.Sprintf("SELECT pos_id, user_id, webhook_url, flask_id FROM %s",
		posTable)
	err := r.db.Select(&terminals, query)
	return terminals, err
}
