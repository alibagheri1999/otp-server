package handlers

import (
	"net/http"
	"strconv"

	"otp-server/internal/application"
	"otp-server/internal/infrastructure/logger"
	"otp-server/internal/interfaces/http/handlers/dto"

	"github.com/gofiber/fiber/v2"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService application.UserServiceInterface
	logger      logger.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService application.UserServiceInterface, logger logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// GetProfile gets the current user's profile
// @Summary Get User Profile
// @Description Get the profile information of the currently authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse "User profile retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/users/profile [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	user, err := h.userService.GetUserByID(c.Context(), userID)
	if err != nil {
		h.logger.Error(c.Context(), "Failed to get user profile", logger.F("error", err), logger.F("user_id", userID))
		return c.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "Failed to get profile",
			Message: err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(dto.UserResponse{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		Name:        user.Name,
		Role:        string(user.Role),
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	})
}

// UpdateProfile updates the current user's profile
// @Summary Update User Profile
// @Description Update the profile information of the currently authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateProfileRequest true "Profile update data"
// @Success 200 {object} dto.UserResponse "Profile updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request data"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/users/profile [put]
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	var req dto.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
	}

	user, err := h.userService.UpdateUserProfile(c.Context(), userID, req.Name)
	if err != nil {
		h.logger.Error(c.Context(), "Failed to update user profile", logger.F("error", err), logger.F("user_id", userID))
		return c.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "Failed to update profile",
			Message: err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(dto.UserResponse{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		Name:        user.Name,
		Role:        string(user.Role),
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	})
}

// SearchUsers is a unified endpoint that handles both search and pagination
// @Summary Get Users Unified
// @Description Get users with optional search and pagination in a single endpoint
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param query query string false "Search query (optional)"
// @Param offset query int false "Pagination offset (default: 0)"
// @Param limit query int false "Pagination limit (default: 10, max: 100)"
// @Success 200 {object} dto.UnifiedUsersResponse "Users retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid parameters"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/users/search [get]
func (h *UserHandler) SearchUsers(c *fiber.Ctx) error {
	// Parse query parameters
	query := c.Query("query", "")

	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "Invalid offset parameter",
			Message: "Offset must be a valid integer",
		})
	}

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "Invalid limit parameter",
			Message: "Limit must be a valid integer",
		})
	}

	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 10
	}

	users, total, err := h.userService.GetUsers(c.Context(), query, offset, limit)
	if err != nil {
		h.logger.Error(c.Context(), "Failed to get unified users", logger.F("error", err), logger.F("query", query), logger.F("offset", offset), logger.F("limit", limit))
		return c.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "Failed to get users",
			Message: err.Error(),
		})
	}

	userResponses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = &dto.UserResponse{
			ID:          user.ID,
			PhoneNumber: user.PhoneNumber,
			Name:        user.Name,
			Role:        string(user.Role),
			IsActive:    user.IsActive,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}
	}

	return c.Status(http.StatusOK).JSON(dto.UnifiedUsersResponse{
		Users: userResponses,
		Total: total,
		Query: query,
		Page: struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}{
			Offset: offset,
			Limit:  limit,
		},
	})
}
