package handlers

import (
	"notification-service/services"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService}
}

func (pc *NotificationHandler) NotificationsRoute(rg *gin.RouterGroup) {

	router := rg.Group("notifications")
	router.POST("/", pc.notificationService.CreateNotification)
	router.PUT("/:id", pc.notificationService.UpdateNotification)
	router.DELETE("/:id", pc.notificationService.DeleteNotification)
	router.GET("/", pc.notificationService.FindNotifications)
	router.GET("/:id", pc.notificationService.FindMessagesByNotificationId)
}
