package main

import (
	"notification-service/inits"
	"notification-service/models"
	"notification-service/services"
	"notification-service/services/interfaces"
)

var (
	logger interfaces.Logger
)

func init() {
	config, logConfig := inits.InitConfig()
	logger = services.NewLoggerZap(logConfig)
	defer logger.Sync()

	inits.ConnectDB(config, logger)
}

func main() {
	inits.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	inits.DB.AutoMigrate(&models.Client{}, &models.Notification{}, &models.Message{})
	logger.Info("migration complete")
}
