package service

type RequestInvoice struct {
	UserID  int    `json:"pos-id"`
	Account string `json:"account"`
	Amount  int    `json:"amount"`
	Message string `json:"message"`
}

type ResponseInvoice struct {
	ClientName string `json:"client-name"`
}

type RequestCheck struct {
	UserID  int      `json:"pos-id"`
	IsToday int      `json:"today"`
	IDs     []string `json:"bills-list"`
}

type RequestCancelInvoice struct {
	UserID int    `json:"pos-id"`
	ID     string `json:"invoice-id"`
}

type RequestCancelPayment struct {
	UserID  int    `json:"pos-id"`
	IsToday int    `json:"today"`
	Amount  int    `json:"amount"`
	ID      string `json:"invoice-id"`
}
