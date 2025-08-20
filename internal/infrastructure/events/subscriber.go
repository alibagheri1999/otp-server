package events

import (
	"context"
	"otp-server/internal/infrastructure/config"
	"otp-server/internal/infrastructure/logger"
	"otp-server/internal/infrastructure/redis"
)

type EventHandler func(ctx context.Context, event *Event) error

type Subscriber struct {
	redisClient *redis.Client
	config      *config.EventsConfig
	logger      logger.Logger
	handlers    map[string][]EventHandler
}

func NewSubscriber(redisClient *redis.Client, cfg *config.EventsConfig, logger logger.Logger) *Subscriber {
	return &Subscriber{
		redisClient: redisClient,
		config:      cfg,
		logger:      logger,
		handlers:    make(map[string][]EventHandler),
	}
}

func (s *Subscriber) Subscribe(ctx context.Context, eventType string, handler EventHandler) error {
	if !s.config.Enabled {
		return nil
	}

	s.handlers[eventType] = append(s.handlers[eventType], handler)
	return nil
}

func (s *Subscriber) Start(ctx context.Context) error {
	if !s.config.Enabled {
		return nil
	}

	pubsub := s.redisClient.Subscribe(ctx, s.config.RedisChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			if msg == nil {
				continue
			}

			var event Event
			if err := event.FromJSON([]byte(msg.Payload)); err != nil {
				continue
			}

			s.handleEvent(ctx, &event)

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Subscriber) handleEvent(ctx context.Context, event *Event) {
	handlers, exists := s.handlers[event.Type]
	if !exists {
		handlers = s.handlers["*"]
	}

	for _, handler := range handlers {
		go func(h EventHandler) {
			if err := h(ctx, event); err != nil {
			}
		}(handler)
	}
}
