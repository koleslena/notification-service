package services

import (
	"net/http"
	"notification-service/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotificationService struct {
	DB       *gorm.DB
	channels *Channels
}

func NewNotificationService(DB *gorm.DB, channels *Channels) *NotificationService {
	return &NotificationService{DB: DB, channels: channels}
}

// @Summary add a new notification
// @ID create-notification
// @Param data body models.Notification true "notification data"
// @Produce json
// @Success 201 {object} models.NotificationRequest
// @Failure 400 {object} string "Error"
// @Failure 502 {object} string "Error"
// @Router /notifications [post]
func (ns *NotificationService) CreateNotification(ctx *gin.Context) {
	var payload *models.NotificationRequest

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	newNotification := models.Notification{
		Text:    payload.Text,
		Filter:  payload.Filter,
		StartAt: payload.StartAt,
		EndAt:   payload.EndAt,
	}

	result := ns.DB.Create(&newNotification)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"Error": result.Error.Error()})
		return
	}
	ns.channels.AddNotification(&newNotification)

	ctx.JSON(http.StatusCreated, newNotification)
}

// @Summary edit a notification by ID
// @ID update-notification
// @Produce json
// @Param id path string true "The ID of a notification"
// @Success 200 {object} models.NotificationRequest
// @Failure 502 {object} string "Error"
// @Failure 404 {object} string "Error"
// @Router /notifications/{id} [put]
func (ns *NotificationService) UpdateNotification(ctx *gin.Context) {
	notificationId := ctx.Param("id")

	var payload *models.NotificationRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"Error": err.Error()})
		return
	}
	var updatedNotification models.Notification
	result := ns.DB.First(&updatedNotification, "id = ?", notificationId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"Error": "No notification with that id exists"})
		return
	}

	id, e := strconv.Atoi(notificationId)
	if e != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"Error": e})
		return
	}
	notificationToUpdate := models.Notification{
		ID:      id,
		Text:    payload.Text,
		Filter:  payload.Filter,
		StartAt: payload.StartAt,
		EndAt:   payload.EndAt,
	}

	ns.DB.Model(&updatedNotification).Updates(notificationToUpdate)

	ns.channels.AddNotification(&notificationToUpdate)

	ctx.JSON(http.StatusOK, updatedNotification)
}

// @Summary delete a notification by ID
// @ID delete-notification
// @Produce json
// @Param id path string true "The ID of a notification"
// @Success 204
// @Failure 404 {object} string "Error"
// @Router /notifications/{id} [delete]
func (ns *NotificationService) DeleteNotification(ctx *gin.Context) {
	notificationId := ctx.Param("id")

	result := ns.DB.Delete(&models.Notification{}, "id = ?", notificationId)

	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"Error": "No notification with that id exists"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// @Summary get all messages for the notification
// @ID find-messages-by-notification-id
// @Produce json
// @Param id path string true "The ID of notification"
// @Success 200 {object} []models.Message
// @Failure 404 {object} string "Error"
// @Router /notifications [get]
func (ns *NotificationService) FindMessagesByNotificationId(ctx *gin.Context) {
	notificationId := ctx.Param("id")
	var page = ctx.DefaultQuery("page", "1")
	var limit = ctx.DefaultQuery("limit", "10")

	intPage, _ := strconv.Atoi(page)
	intLimit, _ := strconv.Atoi(limit)
	offset := (intPage - 1) * intLimit

	var messages []models.Message
	results := ns.DB.Limit(intLimit).Offset(offset).Where("notification_id = ?", notificationId).Find(&messages)
	if results.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"Error": "No messages for that notification exists"})
		return
	}

	ctx.JSON(http.StatusOK, messages)
}

// @Summary get all items in the notification list
// @ID find-notifications
// @Produce json
// @Success 200 {object} []models.Notification
// @Failure 502 {object} string "Error"
// @Router /notifications [get]
func (ns *NotificationService) FindNotifications(ctx *gin.Context) {
	var page = ctx.DefaultQuery("page", "1")
	var limit = ctx.DefaultQuery("limit", "10")

	intPage, _ := strconv.Atoi(page)
	intLimit, _ := strconv.Atoi(limit)
	offset := (intPage - 1) * intLimit

	var notifications []models.Notification
	results := ns.DB.Limit(intLimit).Offset(offset).Find(&notifications)
	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"Error": results.Error})
		return
	}

	ctx.JSON(http.StatusOK, notifications)
}
