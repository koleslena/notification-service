package handlers

import (
	"notification-service/services"

	"github.com/gin-gonic/gin"
)

type ClientHandler struct {
	clientService *services.ClientService
}

func NewClientHandler(clientService *services.ClientService) *ClientHandler {
	return &ClientHandler{clientService}
}

func (pc *ClientHandler) ClientsRoute(rg *gin.RouterGroup) {

	router := rg.Group("clients")
	router.POST("/", pc.clientService.CreateClient)
	router.PUT("/:id", pc.clientService.UpdateClient)
	router.DELETE("/:id", pc.clientService.DeleteClient)
}
