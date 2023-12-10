package handler

import (
	"daemon"
	"daemon/pkg/repository"
	"github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

const TimeOutSec = 10

func (h *Handler) Daemon() {
	posTerminals, err := h.services.GetAllPosTerminals()
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.Println("daemon started")
	running := make(map[int]bool)
	var runningMutex sync.Mutex

	for {
		for _, posTerminal := range posTerminals {
			runningMutex.Lock()

			if _, exists := running[posTerminal.FlaskId]; !exists || !running[posTerminal.FlaskId] {
				running[posTerminal.FlaskId] = true
				runningMutex.Unlock()
				go func(posTerminal daemon.PosTerminal) {
					h.allOperations(posTerminal)
					runningMutex.Lock()
					running[posTerminal.FlaskId] = false
					runningMutex.Unlock()
				}(posTerminal)
			} else {
				runningMutex.Unlock()
			}
		}
		time.Sleep(TimeOutSec * time.Second)
	}
}

func (h *Handler) allOperations(posTerminal daemon.PosTerminal) {
	if err := h.services.SetOldInvoicesToCancel(); err != nil {
		logrus.Error(err)
	}

	invoices, err := h.services.GetInWorkInvoices(posTerminal)
	if err != nil {
		logrus.Error(err)
		return
	}

	var forCheck []string

	for _, invoice := range invoices {
		switch invoice.Status {
		case repository.STATUS1:
			if err = h.services.SendInvoice(invoice, posTerminal); err != nil {
				logrus.Error(err)
			}
		case repository.STATUS2:
			forCheck = append(forCheck, strconv.Itoa(invoice.UUID))
		case repository.STATUS4:
			if err = h.services.CancelInvoice(invoice, posTerminal); err != nil {
				logrus.Error(err)
			}
		case repository.STATUS6:
			if err = h.services.CancelInvoice(invoice, posTerminal); err != nil {
				logrus.Error(err)
			}
		case repository.STATUS10:
			amount, err := h.services.GetInvoiceAmount(invoice)
			if err != nil {
				logrus.Error(err)
				continue
			}
			if err = h.services.CancelPayment(invoice, posTerminal, amount, 1); err != nil {
				logrus.Error(err)
			}
		default:
			continue
		}
	}
	if len(forCheck) > 0 {
		if err = h.services.CheckInvoices(posTerminal, 1, forCheck); err != nil {
			logrus.Error(err)
		}
	}
}
