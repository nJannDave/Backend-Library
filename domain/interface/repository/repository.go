package repository

import (
	"context"
	"time"

	"stmnplibrary/domain/entity"
)

type AuthRepository interface {
	GetPassword(ctx context.Context, nis int) (string, string, error)
	GetId(ctx context.Context, nis int) (int, error)
	RedisGet(ctx context.Context, key string) (any, error)
	RedisSet(ctx context.Context, key string, data any, ttl time.Duration) error
}

type UserRepository interface {
	RateLimiter(ctx context.Context, key string) error
	RedisZR(ctx context.Context, key string, start int, stop int) ([]string, error)
	RedisZS(ctx context.Context, key string, score float64, member interface{}) error
	RedisWtx(ctx context.Context, fn func(ctx context.Context) error) error
	RedisWp(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)
	RedisHGetAll(ctx context.Context, key string, dest any) error
	RedisHSET(ctx context.Context, key string, value any) error
	RedisSet(ctx context.Context, key string, data any, ttl time.Duration) error
	RedisGet(ctx context.Context, key string) (any, error)
	RedisDel(ctx context.Context, key string) error

	WithContext(ctx context.Context, fn func(context.Context) error) error

	Register(ctx context.Context, data *entity.Students) error
	GetNIS(ctx context.Context, nis int) error
	GetEmail(ctx context.Context, email string) error

	GetBooks(ctx context.Context, offset int) ([]entity.Book, error)
	GetBooksByAuthor(ctx context.Context, author string, offset int) ([]entity.Book, error)
	GetBooksByCategory(ctx context.Context, category []string, offset int) ([]entity.Book, error)

	CheckLoan(ctx context.Context, idBook int, idUser int) error
	UpdateLimitLoan(ctx context.Context, id int) error
	UpdateBookStock(ctx context.Context, idBook int) error
	CreateLoan(ctx context.Context, loanData entity.Loan) error
}

type AdminRepository interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error

	GetLDDone(ctx context.Context, offset int) ([]entity.LoanData, error)
	GetLDDont(ctx context.Context, offset int) ([]entity.LoanData, error)

	GetLoanData(ctx context.Context, offset int) ([]entity.LoanData, error)
	GetStudentId(ctx context.Context, studentName string) (int, error)
	GetBookId(ctx context.Context, isbn string) (int, error)
	GetStudentLoan(ctx context.Context, idUser int, idBook int) (entity.LdUpdate, error)
	UpdateTabLoan(ctx context.Context, idUser int, idBook int, sanctions int64, returnedAt time.Time) error
	UpdateStock(ctx context.Context, idBook int) error
	UpdateMaxBook(ctx context.Context, id int) error

	AddCategory(ctx context.Context, data entity.Category) error
	AddBook(ctx context.Context, data *entity.BookData) error
	AddConnections(ctx context.Context, data []entity.Connections) error

	RedisSETNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)
	RedisSet(ctx context.Context, key string, data any, ttl time.Duration) error
	RedisGet(ctx context.Context, key string) (any, error)
	RedisDel(ctx context.Context, key string) error
}
