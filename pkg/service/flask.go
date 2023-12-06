package service

type RequestInvoice struct {
	PosTerminalId int    `json:"pos-id"`
	Account       string `json:"account"`
	Amount        int    `json:"amount"`
	Message       string `json:"message"`
}

type ResponseInvoice struct {
	ClientName string `json:"client-name"`
}

type RequestCheck struct {
	PosTerminalId int      `json:"pos-id"`
	IsToday       int      `json:"today"`
	IDs           []string `json:"bills-list"`
}

type RequestCancelInvoice struct {
	PosTerminalId int    `json:"pos-id"`
	ID            string `json:"invoice-id"`
}

type RequestCancelPayment struct {
	PosTerminalId int    `json:"pos-id"`
	IsToday       int    `json:"today"`
	Amount        int    `json:"amount"`
	ID            string `json:"invoice-id"`
}
