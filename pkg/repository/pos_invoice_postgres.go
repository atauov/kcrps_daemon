package repository

import (
	"daemon"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type PosInvoicePostgres struct {
	db *sqlx.DB
}

func NewPosInvoicePostgres(db *sqlx.DB) *PosInvoicePostgres {
	return &PosInvoicePostgres{db: db}
}

func (r *PosInvoicePostgres) UpdateStatus(invoice daemon.Invoice, status int) error {
	query := fmt.Sprintf(`UPDATE %s SET status=$1 WHERE pos_id=$2 AND uuid = $3`, invoicesTable)
	_, err := r.db.Exec(query, status, invoice.PosID, invoice.UUID)
	logrus.Printf("NEW STATUS = %d", status)
	return err
}

func (r *PosInvoicePostgres) UpdateClientName(invoice daemon.Invoice, clientName string) error {
	query := fmt.Sprintf(`UPDATE %s SET client_name=$1 WHERE pos_id=$2 AND uuid = $3`, invoicesTable)
	_, err := r.db.Exec(query, clientName, invoice.PosID, invoice.UUID)
	return err
}

func (r *PosInvoicePostgres) GetInWorkInvoices(posTerminal daemon.PosTerminal) ([]daemon.Invoice, error) {
	var invoices []daemon.Invoice

	query := fmt.Sprintf("SELECT uuid, status, amount, account, message FROM %s WHERE pos_id=$1 "+
		"AND status IN($2, $3, $4, $5, $6) ORDER BY created_at", invoicesTable)
	err := r.db.Select(&invoices, query, posTerminal.Id, STATUS1, STATUS2, STATUS4, STATUS6, STATUS10)

	return invoices, err
}

func (r *PosInvoicePostgres) GetInvoiceAmount(invoice daemon.Invoice) (int, error) {
	var amount int
	query := fmt.Sprintf(`SELECT amount FROM %s WHERE pos_id=$1 AND uuid=$2`, invoicesTable)
	if err := r.db.Get(&amount, query, invoice.PosID, invoice.UUID); err != nil {
		return 0, err
	}
	return amount, nil
}

func (r *PosInvoicePostgres) GetAllPosTerminals() ([]daemon.PosTerminal, error) {
	var terminals []daemon.PosTerminal
	query := fmt.Sprintf(`SELECT pos_id, user_id, webhook_url, flask_id FROM %s`, posTable)
	err := r.db.Select(&terminals, query)
	return terminals, err
}
