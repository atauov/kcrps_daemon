package service

import (
	"bytes"
	"daemon"
	"daemon/pkg/repository"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
)

const (
	CreateInvoiceURL = "https://ks1.ddns.me/create-invoice"
	CancelInvoiceURL = "https://ks1.ddns.me/cancel-invoice"
	CancelPaymentURL = "https://ks1.ddns.me/cancel-payment"
	CheckInvoicesURL = "https://ks1.ddns.me/check-invoices"
)

type PosInvoiceService struct {
	repo repository.PosInvoice
}

type WebHook struct {
	PosId      string `json:"pos-id"`
	Id         int    `json:"id"`
	Status     int    `json:"status"`
	Account    string `json:"account"`
	ClientName string `json:"client-name"`
	Message    string `json:"message"`
	Amount     int    `json:"amount"`
}

func NewPosInvoiceService(repo repository.PosInvoice) *PosInvoiceService {
	return &PosInvoiceService{repo: repo}
}

func (s *PosInvoiceService) SendInvoice(invoice daemon.Invoice, posTerminal daemon.PosTerminal) error {
	invoiceForFlask := RequestInvoice{
		PosTerminalId: posTerminal.FlaskId,
		Account:       invoice.Account[1:],
		Amount:        invoice.Amount,
		Message:       invoice.Message,
	}

	jsonData, err := json.Marshal(invoiceForFlask)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, CreateInvoiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logrus.Error(err)
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusOK {
		var response ResponseInvoice
		if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return err
		}
		if err = s.repo.UpdateClientName(invoice, response.ClientName); err != nil {
			return err
		}
		if err = s.repo.UpdateStatus(invoice, repository.STATUS2); err != nil {
			return err
		}
		invoice.Status = repository.STATUS2
		sendWebhook(invoice, posTerminal.WebHookURL)

		return nil
	} else if resp.StatusCode == http.StatusNotFound {
		if err = s.repo.UpdateStatus(invoice, repository.STATUS3); err != nil {
			return err
		}
		invoice.Status = repository.STATUS3
		sendWebhook(invoice, posTerminal.WebHookURL)

		return nil
	} else if resp.StatusCode == http.StatusInternalServerError {
		return errors.New("error on pos, please try later")
	}

	return errors.New("unknown error")
}

func (s *PosInvoiceService) CancelInvoice(invoice daemon.Invoice, posTerminal daemon.PosTerminal) error {
	invoiceCancel := RequestCancelInvoice{
		PosTerminalId: posTerminal.FlaskId,
		ID:            strconv.Itoa(invoice.UUID),
	}
	jsonData, err := json.Marshal(invoiceCancel)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, CancelInvoiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("Ошибка при выполнении запроса: %v", err))
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logrus.Error(err)
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusOK {
		var newStatus int
		switch invoice.Status {
		case repository.STATUS4:
			newStatus = repository.STATUS5
		case repository.STATUS6:
			newStatus = repository.STATUS7
		default:
			return err
		}
		if err = s.repo.UpdateStatus(invoice, newStatus); err != nil {
			return err
		}
		invoice.Status = newStatus
		sendWebhook(invoice, posTerminal.WebHookURL)

		return nil
	} else if resp.StatusCode == http.StatusInternalServerError {
		return errors.New("error on pos, please try later")
	}

	return errors.New("unknown error")
}

func (s *PosInvoiceService) CancelPayment(invoice daemon.Invoice, posTerminal daemon.PosTerminal, amount, isToday int) error {

	paymentCancel := RequestCancelPayment{
		PosTerminalId: posTerminal.FlaskId,
		IsToday:       isToday,
		Amount:        amount,
		ID:            strconv.Itoa(invoice.UUID),
	}
	jsonData, err := json.Marshal(paymentCancel)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, CancelPaymentURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("Ошибка при выполнении запроса: %v", err))
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logrus.Error(err)
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusOK {
		if err = s.repo.UpdateStatus(invoice, repository.STATUS11); err != nil {
			return err
		}
		invoice.Status = repository.STATUS11
		sendWebhook(invoice, posTerminal.WebHookURL)

		return nil
	} else if resp.StatusCode == http.StatusInternalServerError {
		return errors.New("error on pos, please try later")
	}

	return errors.New("unknown error")
}

func (s *PosInvoiceService) CheckInvoices(posTerminal daemon.PosTerminal, isToday int, IDs []string) error {
	invoicesForCheck := RequestCheck{
		PosTerminalId: posTerminal.FlaskId,
		IsToday:       isToday,
		IDs:           IDs,
	}
	jsonData, err := json.Marshal(invoicesForCheck)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, CheckInvoicesURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("Ошибка при выполнении запроса: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logrus.Error(err)
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusOK {
		result := make(map[string]int)
		res, _ := io.ReadAll(resp.Body)
		if err = json.Unmarshal(res, &result); err != nil {
			return err
		}
		for k, v := range result {
			uuId, _ := strconv.Atoi(k)

			switch v {
			case 2:
				if err = s.UpdateStatus(daemon.Invoice{UUID: uuId, PosID: posTerminal.Id}, repository.STATUS9); err != nil {
					return err
				}
				sendWebhook(daemon.Invoice{UUID: uuId, Status: repository.STATUS9}, posTerminal.WebHookURL)
			case 1:
				if err = s.UpdateStatus(daemon.Invoice{UUID: uuId, PosID: posTerminal.Id}, repository.STATUS8); err != nil {
					return err
				}
				sendWebhook(daemon.Invoice{UUID: uuId, Status: repository.STATUS8}, posTerminal.WebHookURL)
			}
		}

		return nil
	} else if resp.StatusCode == http.StatusInternalServerError {
		return errors.New("error on pos, please try later")
	}

	return errors.New("unknown error")
}

func (s *PosInvoiceService) UpdateStatus(invoice daemon.Invoice, status int) error {
	return s.repo.UpdateStatus(invoice, status)
}

func (s *PosInvoiceService) UpdateClientName(invoice daemon.Invoice, clientName string) error {
	return s.repo.UpdateClientName(invoice, clientName)
}

func (s *PosInvoiceService) GetInWorkInvoices(posTerminal daemon.PosTerminal) ([]daemon.Invoice, error) {
	return s.repo.GetInWorkInvoices(posTerminal)
}

func (s *PosInvoiceService) GetInvoiceAmount(invoice daemon.Invoice) (int, error) {
	return s.repo.GetInvoiceAmount(invoice)
}

func (s *PosInvoiceService) GetAllPosTerminals() ([]daemon.PosTerminal, error) {
	return s.repo.GetAllPosTerminals()
}

func (s *PosInvoiceService) SetOldInvoicesToCancel() error {
	return s.repo.SetOldInvoicesToCancel()
}
