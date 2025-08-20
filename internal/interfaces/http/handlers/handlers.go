package handlers

import (
	"otp-server/internal/application"
	"otp-server/internal/infrastructure/logger"
)

type Handlers struct {
	AuthHandler *AuthHandler
	UserHandler *UserHandler
	logger      logger.Logger
}

func NewHandlers(services *application.Services, logger logger.Logger) *Handlers {
	return &Handlers{
		AuthHandler: NewAuthHandler(services.AuthService, logger),
		UserHandler: NewUserHandler(services.UserService, logger),
		logger:      logger,
	}
}

// GetLogger returns the logger instance
func (h *Handlers) GetLogger() logger.Logger {
	return h.logger
}
