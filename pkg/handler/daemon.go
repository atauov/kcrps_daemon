package handler

import (
	"github.com/sirupsen/logrus"
	"log"
	"strconv"
	"sync"
	"time"
)

const TimeOutSec = 10

func (h *Handler) Daemon(posIDs []int) {
	logrus.Println("daemon started")
	running := make(map[int]bool)
	var runningMutex sync.Mutex

	for {
		for _, posID := range posIDs {
			// logrus.Printf("unit of pos: %d daemon started", posID)
			runningMutex.Lock()

			if _, exists := running[posID]; !exists || !running[posID] {
				running[posID] = true
				runningMutex.Unlock()
				go func(posID int) {
					h.allOperations(posID)
					runningMutex.Lock()
					running[posID] = false
					runningMutex.Unlock()
				}(posID)
			} else {
				runningMutex.Unlock()
			}
		}
		time.Sleep(TimeOutSec * time.Second)
	}
}

func (h *Handler) allOperations(posID int) {
	invoices, err := h.services.GetInWorkInvoices(posID)
	log.Println(invoices)
	if err != nil {
		logrus.Error(err)
		return
	}

	var forCheck []string

	for _, invoice := range invoices {
		switch invoice.Status {
		case 0:
			if err = h.services.SendInvoice(posID, invoice); err != nil {
				logrus.Error(err)
			}
		case 1:
			forCheck = append(forCheck, strconv.Itoa(invoice.UUID))
		case 3:
			if err = h.services.CancelInvoice(posID, invoice.Id); err != nil {
				logrus.Error(err)
			}
		case 4:
			amount, err := h.services.GetInvoiceAmount(invoice.Id)
			if err != nil {
				logrus.Error(err)
				continue
			}
			if err = h.services.CancelPayment(posID, amount, 1, invoice.Id); err != nil {
				logrus.Error(err)
			}
		}
	}
	if len(forCheck) > 0 {
		if err = h.services.CheckInvoices(posID, 1, forCheck); err != nil {
			logrus.Error(err)
		}
	}
}
