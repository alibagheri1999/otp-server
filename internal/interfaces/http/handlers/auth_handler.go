package handlers

import (
	"net/http"
	"otp-server/lib"

	"otp-server/internal/application"
	"otp-server/internal/infrastructure/logger"
	"otp-server/internal/interfaces/http/handlers/dto"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService application.AuthServiceInterface
	logger      logger.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService application.AuthServiceInterface, logger logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// SendOTP sends OTP to the user's phone number
// @Summary Send OTP
// @Description Send a one-time password (OTP) to the provided phone number for authentication
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.SendOTPRequest true "Send OTP request with phone number"
// @Success 200 {object} dto.SendOTPResponse "OTP sent successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid phone number format"
// @Failure 429 {object} dto.ErrorResponse "Too many OTP requests"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/auth/send-otp [post]
func (h *AuthHandler) SendOTP(c *fiber.Ctx) error {
	var req dto.SendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
	}

	err := lib.ValidatePhoneNumber(req.PhoneNumber)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
	}

	err = h.authService.SendOTP(c.Context(), req.PhoneNumber)
	if err != nil {
		h.logger.Error(c.Context(), "Failed to send OTP", logger.F("error", err), logger.F("phone_number", req.PhoneNumber))
		return c.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "Failed to send OTP",
			Message: err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(dto.SendOTPResponse{
		Message:     "OTP sent successfully",
		PhoneNumber: req.PhoneNumber,
		Timestamp:   c.Get("Date"),
	})
}

// VerifyOTP verifies OTP and returns authentication tokens
// @Summary Verify OTP
// @Description Verify the one-time password (OTP) sent to the user's phone number and return JWT authentication tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.VerifyOTPRequest true "Verify OTP request with phone number and OTP code"
// @Success 200 {object} dto.AuthResponse "Authentication successful - returns access token, refresh token, and user info"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or missing required fields"
// @Failure 401 {object} dto.ErrorResponse "Invalid OTP or expired OTP"
// @Failure 429 {object} dto.ErrorResponse "Too many verification attempts"
// @Failure 500 {object} dto.ErrorResponse "Internal server error during token generation"
// @Router /api/v1/auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var req dto.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
	}

	user, token, err := h.authService.VerifyOTPAndAuthenticate(c.Context(), req.PhoneNumber, req.OTP, req.Name)
	if err != nil {
		h.logger.Error(c.Context(), "Failed to verify OTP", logger.F("error", err), logger.F("phone_number", req.PhoneNumber))
		return c.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "Invalid OTP",
			Message: err.Error(),
		})
	}

	response := dto.AuthResponse{
		Token: token,
		User: dto.AuthUserResponse{
			ID:          user.ID,
			PhoneNumber: user.PhoneNumber,
			Name:        user.Name,
			Role:        string(user.Role),
		},
	}

	return c.Status(http.StatusOK).JSON(response)
}
