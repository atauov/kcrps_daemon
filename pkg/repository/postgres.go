package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

const (
	STATUS1       = iota + 1 //Счет зарегистрирован в системе (надо проверить и выставить в терминале)
	STATUS2                  //Счет проверен и выставлен (надо проверять оплату/отмену)
	STATUS3                  //Счет проверен и не найден аккаунт Каспи (действие не требуется)
	STATUS4                  //Счет ожидает отмены по инициативе продавца (надо отменить в терминале)
	STATUS5                  //Счет отменен по инициативе продавца (действие не требуется)
	STATUS6                  //Счет ожидает отмены по истечению срока оплаты (надо отменить в терминале)
	STATUS7                  //Счет отменен по истечению срока оплаты (действие не требуется)
	STATUS8                  //Счет отклонен по инициативе покупателя (действие не требуется)
	STATUS9                  //Счет успешно оплачен (действие не требуется)
	STATUS10                 //Счет ожидает возврата средств по инициативе продавца (надо отменить в терминале)
	STATUS11                 //Счет возвращен покупателю (действие не требуется)
	posTable      = "pos_terminals"
	invoicesTable = "invoices"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDB(cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.SSLMode))
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
