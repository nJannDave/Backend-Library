package repository

import (
	"stmnplibrary/domain/entity"
	"stmnplibrary/domain/interface/repository"
	"stmnplibrary/controller/repository/utils"

	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type authRepository struct {
	gorm *gorm.DB 
	rds *redis.Client
}

func FnAuthRepository(gorm *gorm.DB, rds *redis.Client) repository.AuthRepository {
	return &authRepository{
		gorm: gorm,
		rds : rds,
	}
}

func (ar *authRepository) validateQuery(result *gorm.DB) error {
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("no data found")
		}
		return fmt.Errorf("internal server error: %w", result.Error)
	}
	return nil
}

func (ar *authRepository) validateExec(result *gorm.DB) error {
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "violates foreign key constraint") {
			return errors.New("id doesn't exist yet")
		}
		return fmt.Errorf("internal server error: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("no data affected")
	}
	return nil
}

func (ar *authRepository) RedisSet(ctx context.Context, key string, data any, ttl time.Duration) error {
	if err := ar.rds.Set(ctx, key, data, ttl).Err(); err != nil {
		return utils.ValidateErrRds(err)
	}
	return nil
}

func (ar *authRepository) RedisGet(ctx context.Context, key string) (any, error) {
	result := ar.rds.Get(ctx, key)
	if result.Err() != nil {
		return nil, utils.ValidateErrRds(result.Err())
	}
	return result, nil
}

func (ar *authRepository) GetPassword(ctx context.Context, nis int) (string, string, error) {
	type data struct {
		Password string `gorm:"column:password"`
		Role string `gorm:"column:role"`
	}
	var d data
	err := ar.gorm.WithContext(ctx).Model(&entity.Students{}).Select("password", "role").Where("nis = ?", nis).Scan(&d)
	if msgErr := ar.validateQuery(err); msgErr != nil {
		return "", "", msgErr
	}
	return d.Password, d.Role, nil
}

func (ar *authRepository) GetId(ctx context.Context, nis int) (int, error) {
	var id int
	err := ar.gorm.WithContext(ctx).Model(&entity.Students{}).Select("id").Where("nis = ?", nis).First(&id)
	if msgErr := ar.validateQuery(err); msgErr != nil {
		return 0, msgErr
	}
	return id, nil
}

