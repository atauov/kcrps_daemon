package handler

import (
	"daemon"
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
	invoices, err := h.services.GetInWorkInvoices(posTerminal.Id)
	if err != nil {
		logrus.Error(err)
		return
	}

	var forCheck []string

	for _, invoice := range invoices {
		switch invoice.Status {
		case 0:
			if err = h.services.SendInvoice(posTerminal, invoice); err != nil {
				logrus.Error(err)
			}
		case 1:
			forCheck = append(forCheck, strconv.Itoa(invoice.UUID))
		case 3:
			if err = h.services.CancelInvoice(posTerminal, invoice.Id); err != nil {
				logrus.Error(err)
			}
		case 4:
			amount, err := h.services.GetInvoiceAmount(invoice.Id)
			if err != nil {
				logrus.Error(err)
				continue
			}
			if err = h.services.CancelPayment(posTerminal, amount, 1, invoice.Id); err != nil {
				logrus.Error(err)
			}
		}
	}
	if len(forCheck) > 0 {
		if err = h.services.CheckInvoices(posTerminal, 1, forCheck); err != nil {
			logrus.Error(err)
		}
	}
}
