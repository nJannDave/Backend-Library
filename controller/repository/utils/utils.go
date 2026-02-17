package utils

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func ValidateErrIDK(err error, msg string) error {
	if err == nil {
		return fmt.Errorf("%s", msg)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return fmt.Errorf("internal server error: %w", err)
}

func ValidateErrRds(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, redis.Nil) {
		return errors.New("no data found")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("operation timed out")
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("operation was canceled by the client")
	}
	return fmt.Errorf("internal server error: %w", err)
}
