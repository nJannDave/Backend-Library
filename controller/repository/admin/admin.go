package repository

import (
	"context"
	"stmnplibrary/domain/entity"
	"stmnplibrary/domain/interface/repository"
	"stmnplibrary/constanta"
	"stmnplibrary/controller/repository/utils"
	"time"
	"fmt"
	"errors"
	"strings"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type adminRepository struct {
	gorm *gorm.DB
	rds  *redis.Client
}

func FnAdminRepository(gorm *gorm.DB, rds *redis.Client) repository.AdminRepository {
	return &adminRepository{
		gorm: gorm,
		rds:  rds,
	}
}

func (ar *adminRepository) getGorm(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(constanta.TX).(*gorm.DB)
	if !ok {
		return ar.gorm
	}
	return tx
}

func (ar *adminRepository) validateQuery(result *gorm.DB) error {
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("no data found")
		}
		return fmt.Errorf("internal server error: %w", result.Error)
	}
	return nil
}

func (ar *adminRepository) validateExec(result *gorm.DB) error {
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

func (ar *adminRepository) RedisSETNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	isNew, err := ar.rds.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return false, utils.ValidateErrRds(err)
	}
	if isNew { 
		return true, nil
	} else {
		return false, nil
	}
}

func (ar *adminRepository) RedisDel(ctx context.Context, key string) error {
	if err := ar.rds.Del(ctx, key).Err(); err != nil {
		return utils.ValidateErrRds(err)
	}
	return nil
}

func (ar *adminRepository) RedisSet(ctx context.Context, key string, data any, ttl time.Duration) error {
	if err := ar.rds.Set(ctx, key, data, ttl).Err(); err != nil {
		return utils.ValidateErrRds(err)
	}
	return nil
}

func (ar *adminRepository) RedisGet(ctx context.Context, key string) (any, error) {
	result := ar.rds.Get(ctx, key)
	if result.Err() != nil {
		return nil, utils.ValidateErrRds(result.Err())
	}
	return result, nil
}

func (ar *adminRepository) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return ar.gorm.Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, constanta.TX, tx)
		return fn(ctx)
	})
}

func (ar *adminRepository) GetLoanData(ctx context.Context, offset int) ([]entity.LoanData, error) {
	var (
		limit    = 35
		loanData = make([]entity.LoanData, 0, limit)
	)
	result := ar.gorm.WithContext(ctx).Model(&entity.LoanData{}).Select(
		"students.name AS student_name",
		"books.name AS book_name",
		"loan.borrow_at",
		"loan.returned_at",
		"loan.must_returned_at",
		"loan.sanctions",
	).Joins("LEFT JOIN students ON students.id = loan.id_user").Joins("LEFT JOIN books on books.id = loan.id_book").Limit(limit).Offset(offset).Scan(&loanData)
	if msgErr := ar.validateQuery(result); msgErr != nil {
		return nil, msgErr
	}
	return loanData, nil
}

func (ar *adminRepository) GetLDDone(ctx context.Context, offset int) ([]entity.LoanData, error) {
	var (
		limit    = 35
		loanData = make([]entity.LoanData, 0, limit)
	)
	result := ar.gorm.WithContext(ctx).Model(&entity.LoanData{}).Select(
		"students.name AS student_name",
		"books.name AS book_name",
		"loan.borrow_at",
		"loan.returned_at",
		"loan.must_returned_at",
		"loan.sanctions",
	).Joins("LEFT JOIN students ON students.id = loan.id_user").Joins("LEFT JOIN books on books.id = loan.id_book").Where("loan.is_returned = ?", true).Limit(limit).Offset(offset).Scan(&loanData)
	if msgErr := ar.validateQuery(result); msgErr != nil {
		return nil, msgErr
	}
	return loanData, nil
}

func (ar *adminRepository) GetLDDont(ctx context.Context, offset int) ([]entity.LoanData, error) {
	var (
		limit    = 35
		loanData = make([]entity.LoanData, 0, limit)
	)
	result := ar.gorm.WithContext(ctx).Model(&entity.LoanData{}).Select(
		"students.name AS student_name",
		"books.name AS book_name",
		"loan.borrow_at",
		"loan.returned_at",
		"loan.must_returned_at",
		"loan.sanctions",
	).Joins("LEFT JOIN students ON students.id = loan.id_user").Joins("LEFT JOIN books on books.id = loan.id_book").Where("loan.is_returned = ?", false).Limit(limit).Offset(offset).Scan(&loanData)
	if msgErr := ar.validateQuery(result); msgErr != nil {
		return nil, msgErr
	}
	return loanData, nil
}

func (ar *adminRepository) AddCategory(ctx context.Context, data entity.Category) error {
	result := ar.gorm.WithContext(ctx).Model(&entity.Category{}).Create(data)
	if msgErr := ar.validateExec(result); msgErr != nil {
		return msgErr
	}
	return nil
}

func (ar *adminRepository) AddBook(ctx context.Context, data *entity.BookData) error {
	gorm := ar.getGorm(ctx)
	result := gorm.WithContext(ctx).Model(&entity.BookData{}).Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).Create(&data)
	if msgErr := ar.validateExec(result); msgErr != nil {
		return msgErr
	}
	return nil
}

func (ar *adminRepository) AddConnections(ctx context.Context, data []entity.Connections) error {
	gorm := ar.getGorm(ctx)
	result := gorm.WithContext(ctx).Model(&entity.Connections{}).Create(data)
	if msgErr := ar.validateExec(result); msgErr != nil {
		return msgErr
	}
	return nil
}

func (ar *adminRepository) GetStudentId(ctx context.Context, studentName string) (int, error) {
	var id int
	result := ar.gorm.WithContext(ctx).Model(&entity.Students{}).Select("id").Where("name = ?", studentName).First(&id)
	if msgErr := ar.validateQuery(result); msgErr != nil {
		return 0, msgErr
	}
	return id, nil
}

func (ar *adminRepository) GetBookId(ctx context.Context, isbn string) (int, error) {
	var id int 
	result := ar.gorm.WithContext(ctx).Model(&entity.Book{}).Select("id").Where("isbn = ?", isbn).First(&id)
	if msgErr := ar.validateQuery(result); msgErr != nil {
		return 0, msgErr
	}
	return id, nil
}

func (ar *adminRepository) GetStudentLoan(ctx context.Context, idUser int, idBook int) (entity.LdUpdate, error) {
	var loanData entity.LdUpdate
	result := ar.gorm.Debug().WithContext(ctx).Table("loan").Model(&entity.LdUpdate{}).Select(
		"must_returned_at", 
		"sanctions", 
		"returned_at",
	).Where("id_book = ?", idBook).Where("id_user = ?", idUser).Where("is_returned = ?", false).Scan(&loanData)
	if msgErr := ar.validateQuery(result); msgErr != nil {
		return entity.LdUpdate{}, msgErr
	}
	return loanData, nil
}

func (ar *adminRepository) UpdateTabLoan(ctx context.Context, idUser int, idBook int, sanctions int64, returnedAt time.Time) error {        
	result := ar.gorm.WithContext(ctx).Model(&entity.LoanData{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id_user = ?", idUser).Where("id_book = ?", idBook).Where("is_returned = ?", false).UpdateColumns(map[string]interface{}{
		"is_returned": true,
		"returned_at": returnedAt,
		"sanctions": gorm.Expr("COALESCE(sanctions, 0) + ?", sanctions),
	})
	if msgErr := ar.validateExec(result); msgErr != nil {
		return msgErr
	}
	return nil
}

func (ar *adminRepository) UpdateStock(ctx context.Context, idBook int) error {
	result := ar.gorm.WithContext(ctx).Model(&entity.Book{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", idBook).UpdateColumn("available_stock", gorm.Expr("available_stock + ?", 1))
	if msgErr := ar.validateExec(result); msgErr != nil {
		return msgErr
	}
	return nil
}

func (ar *adminRepository) UpdateMaxBook(ctx context.Context, id int) error {
	result := ar.gorm.WithContext(ctx).Model(&entity.Students{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).Where("max_book > 0").UpdateColumn("max_book", gorm.Expr("max_book - ?", 1))
	if msgErr := ar.validateExec(result); msgErr != nil {
		return msgErr
	}
	return nil
}