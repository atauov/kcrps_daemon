package service

import (
	"bytes"
	"daemon"
	"daemon/pkg/repository"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
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

func (s *PosInvoiceService) SendInvoice(posTerminal daemon.PosTerminal, invoice daemon.Invoice) error {
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
		if err = s.repo.UpdateClientName(invoice.Id, response.ClientName); err != nil {
			return err
		}
		if err = s.repo.UpdateStatus(invoice.Id, 1, 1); err != nil {
			return err
		}

		sendWebhook(invoice.Id, StatusInvoiceOk, invoice.Account, response.ClientName, posTerminal.WebHookURL)

		return nil
	} else if resp.StatusCode == http.StatusNotFound {
		invoice.InWork = 0

		sendWebhook(invoice.Id, StatusNoAccount, invoice.Account, "unknown", posTerminal.WebHookURL)

		return nil
	} else if resp.StatusCode == http.StatusInternalServerError {
		return errors.New("error on pos, please try later")
	}

	return errors.New("unknown error")
}

func (s *PosInvoiceService) CancelInvoice(posTerminal daemon.PosTerminal, invoiceId int) error {
	invoiceCancel := RequestCancelInvoice{
		PosTerminalId: posTerminal.FlaskId,
		ID:            strconv.Itoa(invoiceId),
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
		if err = s.repo.UpdateStatus(invoiceId, 3, 0); err != nil {
			return err
		}

		sendWebhook(invoiceId, StatusInvoiceCancel, "", "", posTerminal.WebHookURL)

		return nil
	} else if resp.StatusCode == http.StatusInternalServerError {
		return errors.New("error on pos, please try later")
	}

	return errors.New("unknown error")
}

func (s *PosInvoiceService) CancelPayment(posTerminal daemon.PosTerminal, amount, isToday, invoiceId int) error {

	paymentCancel := RequestCancelPayment{
		PosTerminalId: posTerminal.FlaskId,
		IsToday:       isToday,
		Amount:        amount,
		ID:            strconv.Itoa(invoiceId),
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
		if err = s.repo.UpdateStatus(invoiceId, 4, 0); err != nil {
			return err
		}

		sendWebhook(invoiceId, StatusPaymentRefund, "", "", posTerminal.WebHookURL)

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
			uuid, _ := strconv.Atoi(k)
			invoiceId := uuid - 100000
			if err = s.UpdateStatus(invoiceId, v, 0); err != nil {
				return err
			}
			switch v {
			case 2:
				sendWebhook(invoiceId, StatusPaymentOk, "", "", posTerminal.WebHookURL)
			case 1:
				sendWebhook(invoiceId, StatusInvoiceCancel, "", "", posTerminal.WebHookURL)
			}
		}

		return nil
	} else if resp.StatusCode == http.StatusInternalServerError {
		return errors.New("error on pos, please try later")
	}

	return errors.New("unknown error")
}

func (s *PosInvoiceService) UpdateStatus(id, status, inWork int) error {
	return s.repo.UpdateStatus(id, status, inWork)
}

func (s *PosInvoiceService) UpdateClientName(invoiceId int, clientName string) error {
	return s.repo.UpdateClientName(invoiceId, clientName)
}

func (s *PosInvoiceService) GetInWorkInvoices(posTerminalId uuid.UUID) ([]daemon.Invoice, error) {
	return s.repo.GetInWorkInvoices(posTerminalId)
}

func (s *PosInvoiceService) GetInvoiceAmount(invoiceId int) (int, error) {
	return s.repo.GetInvoiceAmount(invoiceId)
}

func (s *PosInvoiceService) GetAllPosTerminals() ([]daemon.PosTerminal, error) {
	return s.repo.GetAllPosTerminals()
}
