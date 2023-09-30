package main

import (
	"context"
	"net/http"
	"notification-service/handlers"
	"notification-service/inits"
	"notification-service/services"
	"notification-service/services/interfaces"

	_ "notification-service/docs"

	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	Config    *inits.Config
	LogConfig *inits.LogConfig

	server    *gin.Engine
	appLogger interfaces.Logger

	Channels    *services.Channels
	SendService *services.SendService

	ClientService *services.ClientService
	ClientHandler *handlers.ClientHandler

	NotificationService *services.NotificationService
	NotificationHandler *handlers.NotificationHandler
)

func init() {
	Config, LogConfig = inits.InitConfig()

	appLogger = services.NewLoggerZap(LogConfig)
	defer appLogger.Sync()

	inits.ConnectDB(Config, appLogger)

	sizes := services.ChannelsSizes{
		MessagesSize:      10,
		ResponsesSize:     10,
		NotificationsSize: 5,
	}
	Channels = services.NewChannels(sizes, appLogger)
	SendService = services.NewSendService(inits.DB, Channels, appLogger, Config)

	NotificationService = services.NewNotificationService(inits.DB, Channels)
	NotificationHandler = handlers.NewNotificationHandler(NotificationService)

	ClientService = services.NewClientService(inits.DB)
	ClientHandler = handlers.NewClientHandler(ClientService)

	server = gin.Default()
}

// @title Notification Service API
// @version 1.0
// @description This is a notification service

// @contact.name koleslena

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8000
// @BasePath /api
// @schemes http
func main() {

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:8000", Config.ClientOrigin}
	corsConfig.AllowCredentials = true

	ctx, _ := context.WithCancel(context.Background())
	go func() {
		if err := SendService.Init(ctx); err != nil {
			appLogger.Infof("sending is stopped with err: %s", err)
		}
	}()
	err := SendService.InitTasks()
	if err != nil {
		appLogger.Infof("init tasks err: %s", err)
	}

	server.Use(cors.New(corsConfig))

	router := server.Group("/api")
	router.GET("/healthchecker", healthCheck())

	ClientHandler.ClientsRoute(router)
	NotificationHandler.NotificationsRoute(router)

	url := ginSwagger.URL("http://localhost:8000/api/swagger/doc.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	appLogger.Error(server.Run(":" + Config.ServerPort))
}

// HealthCheck godoc
// @Summary Show the status of server.
// @Description get the status of server.
// @Tags root
// @Accept */*
// @Produce json
// @Success 200 {object} string "Success"
// @Failure 403 {object} string "Error"
// @Failure 404 {object} string "Error"
// @Router /healthchecker [get]
func healthCheck() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		message := "Welcome to Notification Service"
		ctx.JSON(http.StatusOK, gin.H{"Success": message})
	}
}
