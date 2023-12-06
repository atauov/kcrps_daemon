package handler

import (
	"daemon/pkg/service"
	"sync"
)

type Handler struct {
	services *service.Service
	mutexes  map[int]*sync.Mutex
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{
		services: services,
	}
}
