package service

import (
	"stmnplibrary/domain/interface/repository"
	"stmnplibrary/domain/interface/service"
	"stmnplibrary/security/jwt/claims"
	"stmnplibrary/dto"
	"stmnplibrary/controller/service/utils"
	"stmnplibrary/security/jwt"
	"stmnplibrary/security"

	"context"
	"time"
	"fmt"
)


const keyReTk = "stmnplibrary:accesstoken:id:%d"

type authService struct {
	authRepository repository.AuthRepository
}

func FnAuthService(repository repository.AuthRepository) service.AuthService {
	return &authService {
		authRepository: repository,
	}
}

func (as * authService) Login(ctx context.Context, data dto.Login) (*claims.Token, error) {
	const errIntrnl = "service - login: %w"
	id, err := as.authRepository.GetId(ctx, data.NIS)
	if err != nil {
		return nil, utils.ValidateErrTw(err, errIntrnl)
	}

	var key = fmt.Sprintf(keyReTk, id)

	pH, role, err := as.authRepository.GetPassword(ctx, data.NIS)
	if err != nil {
		return nil, utils.ValidateErrTw(err, errIntrnl)
	}
	if err := security.UnHashPassword(data.Password, pH); err != nil {
		return nil, err
	}
	token, err := token.GenerateToken(id, role)
	if err != nil {
		return nil, fmt.Errorf(errIntrnl, err)
	}

	if err := as.authRepository.RedisSet(ctx, key, []byte(token.RefreshToken), 24*7*time.Hour); err != nil {
		return nil, utils.ValidateErrTw(err, errIntrnl)
	}

	return token, nil
}

func (as *authService) Refresh(ctx context.Context, refreshTkn string) (*claims.Token, error) {
	const errIntrnl = "service - refresh: %w"
	cls, err := token.ValidateToken(refreshTkn)
	if err != nil {
		return nil, err
	}
	var key = fmt.Sprintf(keyReTk, cls.UserId)
	if _, err := as.authRepository.RedisGet(ctx, key); err != nil {
		return nil, utils.ValidateErrTw(err, errIntrnl)
	}
	token, err := token.GenerateToken(cls.UserId, cls.Role)
	if err != nil {
		return nil, err
	}
	if err := as.authRepository.RedisSet(ctx, key, []byte(token.RefreshToken), 24*7*time.Hour); err != nil {
		return nil, utils.ValidateErrTw(err, errIntrnl)
	}
	return token, nil
}