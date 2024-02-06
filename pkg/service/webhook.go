package service

import (
	"bytes"
	"daemon"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
)

func sendWebhook(invoice daemon.Invoice, webhookURL string) {
	jsonWebHook, _ := json.Marshal(WebHook{
		PosId:      invoice.PosID,
		Id:         invoice.UUID,
		Status:     invoice.Status,
		Account:    invoice.Account,
		ClientName: invoice.ClientName,
		Message:    invoice.Message,
		Amount:     invoice.Amount,
	})
	client := &http.Client{}
	resp, _ := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(jsonWebHook))
	if _, err := client.Do(resp); err != nil {
		logrus.Error(err)
	}
}
