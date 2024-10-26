package shared

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type RetryConfig struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
}

func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:     3,
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     2 * time.Second,
		Multiplier:      2.0,
	}
}

func RetryWithBackoff(ctx context.Context, logger *zap.Logger, config *RetryConfig, operation func() error) error {
	var lastErr error
	interval := config.InitialInterval

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := operation(); err == nil {
			return nil
		} else {
			lastErr = err
			logger.Warn("Operation failed, retrying",
				zap.Error(err),
				zap.Int("attempt", attempt+1),
				zap.Duration("next_interval", interval))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(interval):
			}

			interval = time.Duration(float64(interval) * config.Multiplier)
			if interval > config.MaxInterval {
				interval = config.MaxInterval
			}
		}
	}

	return errors.Wrap(lastErr, "max retry attempts reached")
}
