package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"stmnplibrary/domain/entity"
	"stmnplibrary/domain/interface/service"
	"stmnplibrary/dto"
	"stmnplibrary/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setup(t *testing.T) (*mocks.AdminRepository, service.AdminService) {
	repo := mocks.NewAdminRepository(t)
	svc := FnAdminService(repo)
	return repo, svc
}

func TestGetLoanData_All_Methods(t *testing.T) {
	repo, svc := setup(t)
	ctx := context.Background()

	tests := []struct {
		name       string
		methodName string
		page       int
		mockRedis  interface{}
		mockDB     []entity.LoanData
		errDB      error
		expectErr  bool
	}{
		{"Success_From_Redis", "GetLoanData", 1, []dto.LoanData{{StudentName: "Budi"}}, nil, nil, false},
		{"Success_From_DB", "GetLDDone", 2, nil, []entity.LoanData{{StudentName: "Agus"}}, nil, false},
		{"Error_DB_Failure", "GetLDDont", 1, nil, nil, errors.New("db down"), true},
		{"Error_Data_Empty", "GetLoanData", 1, nil, nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := (tt.page - 1) * 35
			var key string

			switch tt.methodName {
			case "GetLoanData":
				key = "stmnplibary:loandata:page:%d"
				repo.On("RedisGet", ctx, mock.Anything).Return(tt.mockRedis, nil).Once()
				if tt.mockRedis == nil {
					repo.On("GetLoanData", ctx, offset).Return(tt.mockDB, tt.errDB).Once()
				}
			case "GetLDDone":
				key = "stmnplibary:loandata:done:page:%d"
				repo.On("RedisGet", ctx, mock.Anything).Return(tt.mockRedis, nil).Once()
				if tt.mockRedis == nil {
					repo.On("GetLDDone", ctx, offset).Return(tt.mockDB, tt.errDB).Once()
				}
			case "GetLDDont":
				key = "stmnplibary:loandata:dont:page:%d"
				repo.On("RedisGet", ctx, mock.Anything).Return(tt.mockRedis, nil).Once()
				if tt.mockRedis == nil {
					repo.On("GetLDDont", ctx, offset).Return(tt.mockDB, tt.errDB).Once()
				}
			}

			if tt.mockRedis == nil && tt.errDB == nil && len(tt.mockDB) > 0 {
				fullKey := fmt.Sprintf(key, tt.page)
				repo.On("RedisSet", ctx, fullKey, mock.Anything, 3*time.Minute).Return(nil).Once()
			}

			var err error
			if tt.methodName == "GetLoanData" {
				_, err = svc.GetLoanData(ctx, tt.page)
			}
			if tt.methodName == "GetLDDone" {
				_, err = svc.GetLDDone(ctx, tt.page)
			}
			if tt.methodName == "GetLDDont" {
				_, err = svc.GetLDDont(ctx, tt.page)
			}

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddCategory_Cases(t *testing.T) {
	repo, svc := setup(t)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		repo.On("AddCategory", ctx, mock.Anything).Return(nil).Once()
		err := svc.AddCategory(ctx, dto.Category{Name: "Horror"})
		assert.NoError(t, err)
	})

	t.Run("Fail_DB_Error", func(t *testing.T) {
		repo.On("AddCategory", ctx, mock.Anything).Return(errors.New("duplicate")).Once()
		err := svc.AddCategory(ctx, dto.Category{Name: "Horror"})
		assert.Error(t, err)
	})
}

func TestAddBook_Cases(t *testing.T) {
	repo, svc := setup(t)
	ctx := context.Background()
	input := dto.BookData{ISBN: "123", Name: "Test", IDCategory: []int{1}}

	t.Run("Success_Complete_Flow", func(t *testing.T) {
		repo.On("WithTx", ctx, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
		repo.On("AddBook", ctx, mock.Anything).Return(nil).Once()
		repo.On("AddConnections", ctx, mock.Anything).Return(nil).Once()

		err := svc.AddBook(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("Fail_AddBook_Inside_Tx", func(t *testing.T) {
		repo.On("WithTx", ctx, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
		repo.On("AddBook", ctx, mock.Anything).Return(errors.New("failed")).Once()

		err := svc.AddBook(ctx, input)
		assert.Error(t, err)
	})
}

func TestConfirm_Cases(t *testing.T) {
	repo, svc := setup(t)
	ctx := context.Background()
	input := dto.Confirm{Student: "A", ISBN: "B"}

	t.Run("Success_Return_On_Time", func(t *testing.T) {
		now := time.Now()
		sanc := int64(0)
		repo.On("GetStudentId", ctx, "A").Return(1, nil).Once()
		repo.On("GetBookId", ctx, "B").Return(2, nil).Once()
		repo.On("GetStudentLoan", ctx, 1, 2).Return(entity.LdUpdate{
			MustReturnedAt: now.Add(24 * time.Hour),
			ReturnedAt:     &now,
			Sanctions:      &sanc,
		}, nil).Once()
		repo.On("WithTx", ctx, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
		repo.On("UpdateTabLoan", ctx, 1, 2, mock.Anything, mock.Anything).Return(nil).Once()
		repo.On("UpdateStock", ctx, 2).Return(nil).Once()

		err := svc.Confirm(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("Fail_Student_Not_Found", func(t *testing.T) {
		repo.On("GetStudentId", ctx, "A").Return(0, errors.New("not found")).Once()
		err := svc.Confirm(ctx, input)
		assert.Error(t, err)
	})

	t.Run("Fail_Update_Stock_Tx", func(t *testing.T) {
		now := time.Now()
		sanc := int64(0)
		repo.On("GetStudentId", ctx, mock.Anything).Return(1, nil).Once()
		repo.On("GetBookId", ctx, mock.Anything).Return(2, nil).Once()
		repo.On("GetStudentLoan", ctx, 1, 2).Return(entity.LdUpdate{MustReturnedAt: now, ReturnedAt: &now, Sanctions: &sanc}, nil).Once()
		repo.On("WithTx", ctx, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
		repo.On("UpdateTabLoan", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		repo.On("UpdateStock", ctx, 2).Return(errors.New("deadlock")).Once()

		err := svc.Confirm(ctx, input)
		assert.Error(t, err)
	})
}