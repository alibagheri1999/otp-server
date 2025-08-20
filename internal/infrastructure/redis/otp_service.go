package redis

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"otp-server/internal/infrastructure/config"
	"otp-server/internal/infrastructure/logger"
	"otp-server/internal/infrastructure/metrics"
)

// OTPService handles OTP generation and validation using Redis
type OTPService struct {
	client       *Client
	logger       logger.Logger
	config       *config.OTPConfig
	eventHandler func(context.Context, string, string) error
	metrics      *metrics.MetricsService
}

// NewOTPService creates a new Redis-based OTP service
func NewOTPService(client *Client, cfg *config.OTPConfig, logger logger.Logger, metricsService *metrics.MetricsService) *OTPService {
	return &OTPService{
		client:  client,
		logger:  logger,
		config:  cfg,
		metrics: metricsService,
	}
}

// SetEventHandler sets the event handler for OTP events
func (s *OTPService) SetEventHandler(handler func(context.Context, string, string) error) {
	s.eventHandler = handler
}

// GenerateOTP generates a new OTP for the given phone number
// Note: Rate limiting is now handled by middleware, not here
func (s *OTPService) GenerateOTP(ctx context.Context, phoneNumber string) (string, error) {
	code, err := s.generateRandomCode(s.config.Length)
	if err != nil {
		return "", err
	}

	otpKey := fmt.Sprintf("%s:%s", s.config.RedisKeyPrefix, phoneNumber)
	err = s.client.Set(ctx, otpKey, code, s.config.Expiry)
	if err != nil {
		return "", err
	}

	if s.metrics != nil {
		s.metrics.RecordOTPGenerated(phoneNumber)
	}

	if s.eventHandler != nil {
		s.eventHandler(ctx, phoneNumber, code)
	}

	return code, nil
}

func (s *OTPService) ValidateOTP(ctx context.Context, phoneNumber, code string) error {
	otpKey := fmt.Sprintf("%s:%s", s.config.RedisKeyPrefix, phoneNumber)
	storedCode, err := s.client.Get(ctx, otpKey)
	if err != nil || storedCode == "" {
		if s.metrics != nil {
			s.metrics.RecordOTPVerified(phoneNumber, false)
		}
		return fmt.Errorf("OTP not found or expired")
	}

	if storedCode != code {
		if s.metrics != nil {
			s.metrics.RecordOTPVerified(phoneNumber, false)
		}
		return fmt.Errorf("invalid OTP code")
	}

	if s.metrics != nil {
		s.metrics.RecordOTPVerified(phoneNumber, true)
	}

	err = s.client.Del(ctx, otpKey)
	if err != nil {
		return err
	}

	return nil
}

func (s *OTPService) CleanupExpiredOTPs(ctx context.Context) error {
	return nil
}

func (s *OTPService) generateRandomCode(length int) (string, error) {
	charset := s.config.CodeCharset
	if charset == "" {
		charset = "0123456789"
	}

	code := ""
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		code += string(charset[num.Int64()])
	}
	return code, nil
}

func (s *OTPService) IsOTPValid(ctx context.Context, phoneNumber string) bool {
	otpKey := fmt.Sprintf("%s:%s", s.config.RedisKeyPrefix, phoneNumber)
	storedCode, err := s.client.Get(ctx, otpKey)
	return err == nil && storedCode != ""
}

func (s *OTPService) GetOTPTTL(ctx context.Context, phoneNumber string) (time.Duration, error) {
	otpKey := fmt.Sprintf("%s:%s", s.config.RedisKeyPrefix, phoneNumber)
	return s.client.TTL(ctx, otpKey)
}
