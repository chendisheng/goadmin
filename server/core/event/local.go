package event

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
)

type LocalBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
	logger   *zap.Logger
}

func NewLocalBus(logger *zap.Logger) *LocalBus {
	return &LocalBus{handlers: make(map[string][]Handler), logger: logger}
}

func (b *LocalBus) Subscribe(topic string, handler Handler) error {
	if b == nil {
		return fmt.Errorf("event bus is not configured")
	}
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return errors.New("event topic is required")
	}
	if handler == nil {
		return errors.New("event handler is required")
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[topic] = append(b.handlers[topic], handler)
	return nil
}

func (b *LocalBus) Publish(ctx context.Context, evt Event) error {
	if b == nil {
		return nil
	}
	if evt == nil {
		return errors.New("event is required")
	}
	topic := strings.TrimSpace(evt.Topic())
	if topic == "" {
		return errors.New("event topic is required")
	}

	handlers := b.snapshot(topic)
	if len(handlers) == 0 {
		return nil
	}

	for _, handler := range handlers {
		h := handler
		go b.dispatch(ctx, topic, evt, h)
	}
	return nil
}

func (b *LocalBus) snapshot(topic string) []Handler {
	b.mu.RLock()
	defer b.mu.RUnlock()
	handlers := b.handlers[topic]
	if len(handlers) == 0 {
		return nil
	}
	return append([]Handler(nil), handlers...)
}

func (b *LocalBus) dispatch(ctx context.Context, topic string, evt Event, handler Handler) {
	defer func() {
		if recovered := recover(); recovered != nil && b.logger != nil {
			b.logger.Error("event handler panic",
				zap.String("topic", topic),
				zap.Any("panic", recovered),
			)
		}
	}()

	if err := handler(ctx, evt); err != nil && b.logger != nil {
		b.logger.Error("event handler failed",
			zap.String("topic", topic),
			zap.Error(err),
		)
	}
}
