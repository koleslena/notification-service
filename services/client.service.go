package services

import (
	"net/http"
	"notification-service/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ClientService struct {
	DB *gorm.DB
}

func NewClientService(DB *gorm.DB) *ClientService {
	return &ClientService{DB}
}

// @Summary add a new client
// @ID create-client
// @Param data body models.Client true "client data"
// @Produce json
// @Success 201 {object} models.ClientRequest
// @Failure 400 {object} string "Error"
// @Failure 502 {object} string "Error"
// @Router /clients [post]
func (pc *ClientService) CreateClient(ctx *gin.Context) {
	var payload *models.ClientRequest

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	newClient := models.Client{
		PhoneNumber: payload.PhoneNumber,
		PhoneCode:   payload.PhoneCode,
		Tag:         payload.Tag,
		TimeZone:    payload.TimeZone,
	}

	result := pc.DB.Create(&newClient)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"Error": result.Error.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, newClient)
}

// @Summary edit a client by ID
// @ID update-client
// @Produce json
// @Param id path string true "The ID of a client"
// @Success 200 {object} models.ClientRequest
// @Failure 502 {object} string "Error"
// @Failure 404 {object} string "Error"
// @Router /clients/{id} [put]
func (pc *ClientService) UpdateClient(ctx *gin.Context) {
	clientId := ctx.Param("id")

	var payload *models.ClientRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"Error": err.Error()})
		return
	}
	var updatedClient models.Client
	result := pc.DB.First(&updatedClient, "id = ?", clientId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"Error": "No client with that id exists"})
		return
	}

	id, e := strconv.Atoi(clientId)
	if e != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"Error": e})
		return
	}
	clientToUpdate := models.Client{
		ID:          id,
		PhoneNumber: payload.PhoneNumber,
		PhoneCode:   payload.PhoneCode,
		Tag:         payload.Tag,
		TimeZone:    payload.TimeZone,
	}

	pc.DB.Model(&updatedClient).Updates(clientToUpdate)

	ctx.JSON(http.StatusOK, updatedClient)
}

// @Summary delete a client by ID
// @ID delete-client
// @Produce json
// @Param id path string true "The ID of a client"
// @Success 204
// @Failure 404 {object} string "Error"
// @Router /clients/{id} [delete]
func (pc *ClientService) DeleteClient(ctx *gin.Context) {
	clientId := ctx.Param("id")

	result := pc.DB.Delete(&models.Client{}, "id = ?", clientId)

	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"Error": "No client with that id exists"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
