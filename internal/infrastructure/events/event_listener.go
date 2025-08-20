package events

import (
	"context"
	"fmt"
	"otp-server/internal/infrastructure/config"
	"otp-server/internal/infrastructure/logger"
)

type EventListener struct {
	logger logger.Logger
	config *config.EventsConfig
}

func NewEventListener(cfg *config.EventsConfig, logger logger.Logger) *EventListener {
	return &EventListener{
		logger: logger,
		config: cfg,
	}
}

func (el *EventListener) HandleOTPEvent(ctx context.Context, event *Event) error {
	if event.Type == el.config.EventTypes.OTPGenerated.Name {
		phoneNumber, _ := event.Payload["phone_number"].(string)
		otpCode, _ := event.Payload["otp_code"].(string)

		fmt.Printf("OTP Generated: %s for %s\n", otpCode, phoneNumber)

		el.logger.Info(ctx, "OTP event processed",
			logger.F("event_type", event.Type),
			logger.F("phone_number", phoneNumber),
			logger.F("event_id", event.ID))
	}

	if event.Type == el.config.EventTypes.OTPVerified.Name {
		phoneNumber, _ := event.Payload["phone_number"].(string)
		userID, _ := event.Payload["user_id"].(int)

		fmt.Printf("OTP Verified for %s, User ID: %d\n", phoneNumber, userID)

		el.logger.Info(ctx, "OTP verified event processed",
			logger.F("event_type", event.Type),
			logger.F("phone_number", phoneNumber),
			logger.F("user_id", userID),
			logger.F("event_id", event.ID))
	}

	return nil
}

func (el *EventListener) HandleUserEvent(ctx context.Context, event *Event) error {
	if event.Type == el.config.EventTypes.UserCreated.Name {
		phoneNumber, _ := event.Payload["phone_number"].(string)
		userID, _ := event.Payload["user_id"].(int)

		fmt.Printf("User Created: ID %d, Phone: %s\n", userID, phoneNumber)

		el.logger.Info(ctx, "User created event processed",
			logger.F("event_type", event.Type),
			logger.F("phone_number", phoneNumber),
			logger.F("user_id", userID),
			logger.F("event_id", event.ID))
	}

	if event.Type == el.config.EventTypes.UserLoggedIn.Name {
		phoneNumber, _ := event.Payload["phone_number"].(string)
		userID, _ := event.Payload["user_id"].(int)

		fmt.Printf("User Logged In: ID %d, Phone: %s\n", userID, phoneNumber)

		el.logger.Info(ctx, "User login event processed",
			logger.F("event_type", event.Type),
			logger.F("phone_number", phoneNumber),
			logger.F("user_id", userID),
			logger.F("event_id", event.ID))
	}

	return nil
}

func (el *EventListener) HandleRateLimitEvent(ctx context.Context, event *Event) error {
	if event.Type == el.config.EventTypes.RateLimited.Name {
		endpoint, _ := event.Payload["endpoint"].(string)
		identifier, _ := event.Payload["identifier"].(string)

		fmt.Printf("Rate Limit Exceeded: %s for %s\n", endpoint, identifier)

		el.logger.Warn(ctx, "Rate limit event processed",
			logger.F("event_type", event.Type),
			logger.F("endpoint", endpoint),
			logger.F("identifier", identifier),
			logger.F("event_id", event.ID))
	}

	return nil
}

func (el *EventListener) HandleAllEvents(ctx context.Context, event *Event) error {
	el.logger.Debug(ctx, "Event received",
		logger.F("event_type", event.Type),
		logger.F("event_id", event.ID),
		logger.F("payload", event.Payload))

	switch event.Type {
	case el.config.EventTypes.OTPGenerated.Name, el.config.EventTypes.OTPVerified.Name:
		return el.HandleOTPEvent(ctx, event)
	case el.config.EventTypes.UserCreated.Name, el.config.EventTypes.UserLoggedIn.Name:
		return el.HandleUserEvent(ctx, event)
	case el.config.EventTypes.RateLimited.Name:
		return el.HandleRateLimitEvent(ctx, event)
	default:
		el.logger.Debug(ctx, "Unknown event type", logger.F("event_type", event.Type))
	}

	return nil
}

// StartEventListener starts listening to events
func (el *EventListener) StartEventListener(ctx context.Context, eventService *EventService) error {
	// Subscribe to all event types
	if err := eventService.Subscribe(ctx, "*", el.HandleAllEvents); err != nil {
		return err
	}

	// Start the event service
	return eventService.Start(ctx)
}
