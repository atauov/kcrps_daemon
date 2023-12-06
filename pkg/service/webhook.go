package service

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
)

func sendWebhook(invoiceId int, status string, account string, clientName string, webhookURL string) {
	jsonWebHook, _ := json.Marshal(WebHook{
		Id:         invoiceId,
		Status:     status,
		Account:    account,
		ClientName: clientName,
	})
	client := &http.Client{}
	resp, _ := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(jsonWebHook))
	if _, err := client.Do(resp); err != nil {
		logrus.Error(err)
	}
}
