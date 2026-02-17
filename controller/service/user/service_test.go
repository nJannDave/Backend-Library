package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"stmnplibrary/domain/entity"
	"stmnplibrary/domain/interface/service"
	"stmnplibrary/dto"
	"stmnplibrary/constanta"
	"stmnplibrary/mocks"
	token "stmnplibrary/security/jwt"
	"stmnplibrary/security"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUser(t *testing.T) (*mocks.UserRepository, service.UserService) {
	repo := mocks.NewUserRepository(t)
	svc := FnUserService(repo)
	return repo, svc
}

func TestRegister(t *testing.T) {
	repo, svc := setupUser(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		input     *dto.Students
		mockSetup func()
		expectMsg bool
		expectErr bool
	}{
		{
			name: "Success",
			input: &dto.Students{
				NIS: 111, Name: "User", Email: "u@m.com", 
				PhoneNumber: "08123456789", Password: "pass", 
				Major: "SIJA", Class: "XI",
			},
			mockSetup: func() {
				repo.On("GetNIS", ctx, 111).Return(nil).Once()
				repo.On("GetEmail", ctx, "u@m.com").Return(nil).Once()
				repo.On("Register", ctx, mock.Anything).Return(nil).Once()
			},
			expectMsg: false,
			expectErr: false,
		},
		{
			name: "Fail_Validation",
			input: &dto.Students{NIS: 999, Email: "fail@m.com", PhoneNumber: "123"},
			mockSetup: func() {
				repo.On("GetNIS", ctx, 999).Return(nil).Once()
				repo.On("GetEmail", ctx, "fail@m.com").Return(nil).Once()
			},
			expectMsg: true,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			msg, err := svc.Register(ctx, tt.input)
			if tt.expectErr { assert.Error(t, err) } else { assert.NoError(t, err) }
			if tt.expectMsg { assert.NotEmpty(t, msg) } else { assert.Nil(t, msg) }
		})
	}
}

func TestLogin(t *testing.T) {
	repo, svc := setupUser(t)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		password := "password123"
		hashedPassword, _ := security.HashPassword(password)

		repo.On("GetId", ctx, 123).Return(1, nil).Once()
		repo.On("GetPassword", ctx, 123).Return(hashedPassword, nil).Once()
		repo.On("RedisSet", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		res, err := svc.Login(ctx, dto.Login{NIS: 123, Password: password})
		
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}

func TestGetBooks(t *testing.T) {
	repo, svc := setupUser(t)
	ctx := context.Background()

	t.Run("Cache_Hit", func(t *testing.T) {
		repo.On("RedisZR", ctx, mock.Anything, 0, 34).Return([]string{"1"}, nil).Once()
		repo.On("RedisWp", ctx, mock.Anything).Return([]dto.Books{{ID: 1}}, nil).Once()

		res, err := svc.GetBooks(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("Cache_Miss_DB_Success", func(t *testing.T) {
		repo.On("RedisZR", ctx, mock.Anything, 0, 34).Return([]string{}, nil).Once()
		repo.On("RedisWp", ctx, mock.Anything).Return([]dto.Books{}, nil).Once()
		repo.On("GetBooks", ctx, 0).Return([]entity.Book{{ID: 10}}, nil).Once()
		repo.On("RedisWp", ctx, mock.Anything).Return(nil, nil).Once()

		res, err := svc.GetBooks(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, 10, res[0].ID)
	})
}

func TestGetBooksByAuthor(t *testing.T) {
	repo, svc := setupUser(t)
	ctx := context.Background()

	t.Run("Success_DB", func(t *testing.T) {
		repo.On("RedisZR", ctx, mock.Anything, 0, 34).Return([]string{}, nil).Once()
		repo.On("RedisWp", ctx, mock.Anything).Return([]dto.Books{}, nil).Once()
		repo.On("GetBooksByAuthor", ctx, "Author A", 0).Return([]entity.Book{{ID: 5}}, nil).Once()
		repo.On("RedisWp", ctx, mock.Anything).Return(nil, nil).Once()

		res, err := svc.GetBooksByAuthor(ctx, "Author A", 1)
		assert.NoError(t, err)
		assert.Equal(t, 5, res[0].ID)
	})
}

func TestGetBooksByCategory(t *testing.T) {
	repo, svc := setupUser(t)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		repo.On("RedisWp", ctx, mock.Anything).Return([]dto.Books{}, nil).Once()
		repo.On("GetBooksByCategory", ctx, []string{"Sains"}, 0).Return([]entity.Book{{ID: 9}}, nil).Once()
		repo.On("RedisWp", ctx, mock.Anything).Return(nil, nil).Once()

		res, err := svc.GetBooksByCategory(ctx, []string{"Sains"}, 1)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	})
}

func TestLoan(t *testing.T) {
	repo, svc := setupUser(t)
	ctx := context.WithValue(context.Background(), constanta.UI, 1)

	tests := []struct {
		name      string
		input     dto.Loan
		mockSetup func()
		expectErr bool
	}{
		{
			name: "Success",
			input: dto.Loan{ID: 10, ReturnedAt: time.Now().AddDate(0, 0, 2).Format("02-01-2006")},
			mockSetup: func() {
				repo.On("WithContext", ctx, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				}).Once()
				repo.On("CheckLoan", ctx, 10, 1).Return(nil).Once()
				repo.On("CreateLoan", ctx, mock.Anything).Return(nil).Once()
				repo.On("UpdateBookStock", ctx, 10).Return(nil).Once()
				repo.On("UpdateLimitLoan", ctx, 1).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Fail_CheckLoan",
			input: dto.Loan{ID: 10, ReturnedAt: "01-01-2025"},
			mockSetup: func() {
				repo.On("WithContext", ctx, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				}).Once()
				repo.On("CheckLoan", ctx, 10, 1).Return(errors.New("limit reached")).Once()
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := svc.Loan(ctx, tt.input)
			if tt.expectErr { assert.Error(t, err) } else { assert.NoError(t, err) }
		})
	}
}

func TestLogout(t *testing.T) {
	repo, svc := setupUser(t)
	ctx := context.WithValue(context.Background(), constanta.UI, 1)
	ctx = context.WithValue(ctx, constanta.TokenA, "token-string")

	t.Run("Success", func(t *testing.T) {
		repo.On("RedisWtx", ctx, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
		repo.On("RedisDel", ctx, mock.Anything).Return(nil).Once()
		repo.On("RedisSet", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.Logout(ctx)
		assert.NoError(t, err)
	})
}

func TestRefresh(t *testing.T) {
	repo, svc := setupUser(t)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		repo.On("RedisGet", ctx, mock.Anything).Return([]byte("valid-refresh-token"), nil).Once()
		repo.On("RedisSet", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		validToken, _ := token.GenerateToken(1)
		res, err := svc.Refresh(ctx, validToken.RefreshToken)

		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("Fail_Invalid_Token", func(t *testing.T) {
		res, err := svc.Refresh(ctx, "invalid")
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestCheckAccTkn(t *testing.T) {
	repo, svc := setupUser(t)
	ctx := context.Background()
	t.Run("Blacklisted", func(t *testing.T) {
		repo.On("RedisGet", ctx, mock.Anything).Return([]byte("blacklisted"), nil).Once()
		err := svc.CheckAccTkn(ctx, "tkn"); assert.Error(t, err)
	})
	t.Run("Clean", func(t *testing.T) {
		repo.On("RedisGet", ctx, mock.Anything).Return(nil, nil).Once()
		err := svc.CheckAccTkn(ctx, "tkn"); assert.NoError(t, err)
	})
}