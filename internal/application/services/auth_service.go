package services

import (
	"context"
	"fmt"
	"time"

	"otp-server/internal/domain/entities"
	"otp-server/internal/domain/repositories"
	"otp-server/internal/infrastructure/logger"
	"otp-server/internal/infrastructure/metrics"
	"otp-server/internal/infrastructure/redis"

	"github.com/golang-jwt/jwt/v5"
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo   repositories.UserRepository
	otpService *redis.OTPService
	logger     logger.Logger
	jwtSecret  string
	metrics    *metrics.MetricsService
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repositories.UserRepository, otpService *redis.OTPService, logger logger.Logger, jwtSecret string, metricsService *metrics.MetricsService) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		otpService: otpService,
		logger:     logger,
		jwtSecret:  jwtSecret,
		metrics:    metricsService,
	}
}

func (s *AuthService) SendOTP(ctx context.Context, phoneNumber string) error {
	if !s.isValidPhoneNumber(phoneNumber) {
		return fmt.Errorf("invalid phone number format")
	}

	_, err := s.otpService.GenerateOTP(ctx, phoneNumber)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) VerifyOTPAndAuthenticate(ctx context.Context, phoneNumber, otpCode, name string) (*entities.User, string, error) {
	if err := s.otpService.ValidateOTP(ctx, phoneNumber, otpCode); err != nil {
		return nil, "", err
	}

	user, err := s.userRepo.GetByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		user = entities.NewUser(phoneNumber, name)
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, "", fmt.Errorf("failed to create user")
		}

		if s.metrics != nil {
			s.metrics.RecordUserRegistration(user.ID, phoneNumber)
		}
	} else {
		user.UpdateLastSeen()
		if err := s.userRepo.Update(ctx, user); err != nil {
		}

		if s.metrics != nil {
			s.metrics.RecordUserLogin(user.ID, phoneNumber)
		}
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate authentication token")
	}

	return user, token, nil
}

func (s *AuthService) GetUserFromToken(tokenString string) (*entities.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int(claims["user_id"].(float64))
		phoneNumber := claims["phone_number"].(string)

		user, err := s.userRepo.GetByID(context.Background(), userID)
		if err != nil {
			return nil, fmt.Errorf("user not found")
		}

		if user.PhoneNumber != phoneNumber {
			return nil, fmt.Errorf("token mismatch")
		}

		return user, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (s *AuthService) generateJWT(user *entities.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":      user.ID,
		"phone_number": user.PhoneNumber,
		"name":         user.Name,
		"role":         user.Role,
		"exp":          time.Now().Add(24 * time.Hour).Unix(),
		"iat":          time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) isValidPhoneNumber(phoneNumber string) bool {
	if len(phoneNumber) < 10 || len(phoneNumber) > 15 {
		return false
	}

	if phoneNumber[0] == '+' {
		for _, char := range phoneNumber[1:] {
			if char < '0' || char > '9' {
				return false
			}
		}
		return true
	}

	return false
}
