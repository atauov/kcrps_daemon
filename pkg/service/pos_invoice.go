package service

import (
	"bytes"
	"daemon"
	"daemon/pkg/repository"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
)

const (
	CreateInvoiceURL    = "http://localhost:8080/create-invoice"
	CancelInvoiceURL    = "http://localhost:8080/cancel-invoice"
	CancelPaymentURL    = "http://localhost:8080/cancel-payment"
	CheckInvoicesURL    = "http://localhost:8080/check-invoices"
	StatusInvoiceOk     = "Invoice successful sent"
	StatusNoAccount     = "No kaspi account on number"
	StatusPaymentOk     = "Payment successful"
	StatusInvoiceCancel = "Invoice has been cancelled"
	StatusPaymentRefund = "Refund successful"
)

type PosInvoiceService struct {
	repo repository.PosInvoice
}

type WebHook struct {
	Id         int    `json:"id"`
	Status     string `json:"status"`
	Account    string `json:"account"`
	ClientName string `json:"client-name"`
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
		if err = s.repo.UpdateStatus(invoice, 1); err != nil {
			return err
		}

		sendWebhook(invoice.UUID, StatusInvoiceOk, invoice.Account, response.ClientName, posTerminal.WebHookURL)

		return nil
	} else if resp.StatusCode == http.StatusNotFound {
		if err = s.repo.UpdateStatus(invoice, repository.STATUS3); err != nil {
			return err
		}

		sendWebhook(invoice.UUID, StatusNoAccount, invoice.Account, "unknown", posTerminal.WebHookURL)

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
		logrus.Fatalf("Ошибка при выполнении запроса: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logrus.Error(err)
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusOK {
		if err = s.repo.UpdateStatus(invoice, repository.STATUS5); err != nil {
			return err
		}

		sendWebhook(invoice.UUID, StatusInvoiceCancel, "", "", posTerminal.WebHookURL)

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
		logrus.Fatalf("Ошибка при выполнении запроса: %v", err)
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

		sendWebhook(invoice.UUID, StatusPaymentRefund, invoice.Account, invoice.ClientName, posTerminal.WebHookURL)

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
				if err = s.UpdateStatus(daemon.Invoice{UUID: uuId}, repository.STATUS9); err != nil {
					return err
				}
				sendWebhook(uuId, StatusPaymentOk, "", "", posTerminal.WebHookURL)
			case 1:
				if err = s.UpdateStatus(daemon.Invoice{UUID: uuId}, repository.STATUS8); err != nil {
					return err
				}
				sendWebhook(uuId, StatusInvoiceCancel, "", "", posTerminal.WebHookURL)
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

func (s *PosInvoiceService) GetInWorkInvoices(posTerminalId uuid.UUID) ([]daemon.Invoice, error) {
	return s.repo.GetInWorkInvoices(posTerminalId)
}

func (s *PosInvoiceService) GetInvoiceAmount(invoice daemon.Invoice) (int, error) {
	return s.repo.GetInvoiceAmount(invoice)
}

func (s *PosInvoiceService) GetAllPosTerminals() ([]daemon.PosTerminal, error) {
	return s.repo.GetAllPosTerminals()
}
