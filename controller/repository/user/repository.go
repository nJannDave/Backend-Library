package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"stmnplibrary/constanta"
	"stmnplibrary/controller/repository/utils"
	"stmnplibrary/domain/entity"
	"stmnplibrary/domain/interface/repository"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userRepository struct {
	gorm *gorm.DB
	rds  *redis.Client
}

func FnUserRepository(gorm *gorm.DB, rds *redis.Client) repository.UserRepository {
	return &userRepository{
		gorm: gorm,
		rds:  rds,
	}
}

func (ur *userRepository) getRDS(ctx context.Context, opt string) redis.Cmdable {
	if opt == "tx" {
		tx, ok := ctx.Value(constanta.RTX).(*redis.Pipeliner)
		if !ok {
			return ur.rds
		}
		return *tx
	}
	batch, ok := ctx.Value(constanta.WP).(redis.Pipeliner)
	if !ok {
		return ur.rds
	}
	return batch
}

func (ur *userRepository) getDb(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(constanta.TX).(*gorm.DB)
	if !ok {
		return ur.gorm
	}
	return tx
}

func (ur *userRepository) validateQuery(result *gorm.DB) error {
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("no data found")
		}
		return fmt.Errorf("internal server error: %w", result.Error)
	}
	return nil
}

func (ur *userRepository) validateExec(result *gorm.DB) error {
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

func (ur *userRepository) RateLimiter(ctx context.Context, key string) error {
	limit, _ := strconv.Atoi(os.Getenv("LIMIT"))
	limiter := redis_rate.NewLimiter(ur.rds)
	res, err := limiter.Allow(ctx, key, redis_rate.PerMinute(limit))
	if err != nil {
		return fmt.Errorf("internal server error: %w", err)
	}
	if res.Allowed == 0 {
		return fmt.Errorf("too many request. Try again in: %v", res.RetryAfter)
	}
	return nil
}

func (ur *userRepository) RedisZS(ctx context.Context, key string, score float64, member interface{}) error {
	if err := ur.rds.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}); err != nil {
		return utils.ValidateErrRds(err.Err())
	}
	if err := ur.rds.Expire(ctx, key, 5*time.Minute); err != nil {
		return utils.ValidateErrRds(err.Err())
	}
	return nil
}

func (ur *userRepository) RedisZR(ctx context.Context, key string, start int, stop int) ([]string, error) {
	result := ur.rds.ZRange(ctx, key, int64(start), int64(stop))
	if msgErr := utils.ValidateErrRds(result.Err()); msgErr != nil {
		return nil, msgErr
	}
	return result.Val(), nil
}

func (ur *userRepository) RedisWp(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	var pipe = ur.rds.Pipeline()
	ctx = context.WithValue(ctx, constanta.WP, pipe)
	result, err := fn(ctx)
	if err != nil {
		return nil, utils.ValidateErrRds(err)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("internal server error: %w", err)
	}
	return result, nil
}

func (ur *userRepository) RedisWtx(ctx context.Context, fn func(ctx context.Context) error) error {
	var txPipe = ur.rds.TxPipeline()
	ctx = context.WithValue(ctx, constanta.RTX, txPipe)
	defer func() {
		if r := recover(); r != nil {
			txPipe.Discard()
		}
	}()
	if err := fn(ctx); err != nil {
		txPipe.Discard()
		return utils.ValidateErrRds(err)
	}
	if _, err := txPipe.Exec(ctx); err != nil {
		return fmt.Errorf("internal server error: %w", err)
	}
	return nil
}

func (ur *userRepository) RedisHGetAll(ctx context.Context, key string, dest any) error {
	var rds = ur.getRDS(ctx, "batch")
	if err := rds.HGetAll(ctx, key).Scan(&dest); err != nil {
		return utils.ValidateErrRds(err)
	}
	return nil
}

func (ur *userRepository) RedisHSET(ctx context.Context, key string, value any) error {
	var rds = ur.getRDS(ctx, "batch")
	if err := rds.HSet(ctx, key, value, 5*time.Minute).Err(); err != nil {
		return utils.ValidateErrRds(err)
	}
	return nil
}

func (ur *userRepository) RedisDel(ctx context.Context, key string) error {
	var rds = ur.getRDS(ctx, "tx")
	if err := rds.Del(ctx, key).Err(); err != nil {
		return utils.ValidateErrRds(err)
	}
	return nil
}

func (ur *userRepository) RedisSet(ctx context.Context, key string, data any, ttl time.Duration) error {
	var rds = ur.getRDS(ctx, "tx")
	if err := rds.Set(ctx, key, data, ttl).Err(); err != nil {
		return utils.ValidateErrRds(err)
	}
	return nil
}

func (ur *userRepository) RedisGet(ctx context.Context, key string) (any, error) {
	result := ur.rds.Get(ctx, key)
	if result.Err() != nil {
		return nil, utils.ValidateErrRds(result.Err())
	}
	return result, nil
}

func (ur *userRepository) WithContext(ctx context.Context, fn func(context.Context) error) error {
	return ur.gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, constanta.TX, tx)
		return fn(ctx)
	})
}

func (ur *userRepository) Register(ctx context.Context, data *entity.Students) error {
	result := ur.gorm.WithContext(ctx).Create(&data)
	if result.Error != nil {
		return fmt.Errorf("failed save data: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no data saved")
	}
	return nil
}

func (ur *userRepository) GetNIS(ctx context.Context, nis int) error {
	var result entity.Students
	err := ur.gorm.WithContext(ctx).Select("nis").Where("nis = ?", nis).First(&result).Error
	if msgErr := utils.ValidateErrIDK(err, "nis already registered"); msgErr != nil {
		return msgErr
	}
	return nil
}

func (ur *userRepository) GetEmail(ctx context.Context, email string) error {
	var result entity.Students
	err := ur.gorm.WithContext(ctx).Select("email").Where("email = ?", email).First(&result).Error
	if msgErr := utils.ValidateErrIDK(err, "email already used"); msgErr != nil {
		return msgErr
	}
	return nil
}

func (ur *userRepository) GetBooks(ctx context.Context, offset int) ([]entity.Book, error) {
	var (
		limit = 35
		books []entity.Book
	)
	result := ur.gorm.WithContext(ctx).Select("id", "name", "author", "publisher", "description", "available_stock").Preload("Categories", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Limit(limit).Offset(offset).Find(&books)
	if msgErr := ur.validateQuery(result); msgErr != nil {
		return nil, msgErr
	}
	return books, nil
}

func (ur *userRepository) GetBooksByAuthor(ctx context.Context, author string, offset int) ([]entity.Book, error) {
	var (
		limit = 35
		books []entity.Book
		where = "%" + author + "%"
	)
	result := ur.gorm.WithContext(ctx).Select("books.id", "books.name", "books.author", "books.publisher", "books.description", "books.available_stock").Preload("Categories").Joins("JOIN connections ON connections.id_book = books.id").Joins("JOIN categories ON categories.id = connections.id_category").Where("books.author ILIKE ?", where).Limit(limit).Offset(offset).Distinct().Find(&books)
	if msgErr := ur.validateQuery(result); msgErr != nil {
		return nil, msgErr
	}
	return books, nil
}

func (ur *userRepository) GetBooksByCategory(ctx context.Context, category []string, offset int) ([]entity.Book, error) {
	var (
		books []entity.Book
		limit = 35
	)
	result := ur.gorm.WithContext(ctx).Select("books.id", "books.name", "books.author", "books.publisher", "books.description", "books.available_stock").Preload("Categories").Joins("JOIN connections ON connections.id_book = books.id").Joins("JOIN categories ON categories.id = connections.id_category").Where("categories.name IN ?", category).Limit(limit).Offset(offset).Distinct().Find(&books)
	if msgErr := ur.validateQuery(result); msgErr != nil {
		return nil, msgErr
	}
	return books, nil
}

func (ur *userRepository) CheckLoan(ctx context.Context, idBook int, idUser int) error {
	var test entity.Loan
	const isReturn = false
	result := ur.gorm.WithContext(ctx).Select("id_user", "id_book").Where("id_user = ?", idUser).Where("id_book = ?", idBook).Where("is_returned = ?", false).First(&test)
	if msgErr := utils.ValidateErrIDK(result.Error, "you have already borrowed that book"); msgErr != nil {
		return msgErr
	}
	return nil
}

func (ur *userRepository) UpdateLimitLoan(ctx context.Context, id int) error {
	result := ur.getDb(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Model(&entity.Students{}).WithContext(ctx).Where("id = ?", id).Where("max_book < ?", 3).UpdateColumn("max_book", gorm.Expr("max_book + ?", 1))
	if msgErr := ur.validateExec(result); msgErr != nil {
		return msgErr
	}
	return nil
}

func (ur *userRepository) UpdateBookStock(ctx context.Context, idBook int) error {
	result := ur.gorm.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Model(&entity.Book{}).Where("id = ?", idBook).Where("available_stock > ?", 0).UpdateColumn("available_stock", gorm.Expr("available_stock - ?", 1))
	if msgErr := ur.validateExec(result); msgErr != nil {
		return msgErr
	}
	return nil
}

func (ur *userRepository) CreateLoan(ctx context.Context, loanData entity.Loan) error {
	result := ur.getDb(ctx).WithContext(ctx).Create(&loanData)
	if msgErr := ur.validateExec(result); msgErr != nil {
		return msgErr
	}
	return nil
}
