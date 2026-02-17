package service

import (
	"context"

	"stmnplibrary/dto"
	"stmnplibrary/security/jwt/claims"
)

type AuthService interface {
	Login(ctx context.Context, data dto.Login) (*claims.Token, error)
	Refresh(ctx context.Context, refreshTkn string) (*claims.Token, error)
}

type UserService interface {
	RateLimiter(ctx context.Context, ip string) error
	CheckAccTkn(ctx context.Context, acc string) error
	
	Register(ctx context.Context, data *dto.Students) ([]string, error)
	Logout(ctx context.Context) error

	GetBooks(ctx context.Context, page int) ([]dto.Books, error)
	GetBooksByAuthor(ctx context.Context, author string, page int) ([]dto.Books, error)
	GetBooksByCategory(ctx context.Context, category []string, page int) ([]dto.Books, error)
	Loan(ctx context.Context, loanInfo dto.Loan) error
}

type AdminService interface {
	GetLoanData(ctx context.Context, page int) ([]dto.LoanData, error)
	GetLDDone(ctx context.Context, page int) ([]dto.LoanData, error)
	GetLDDont(ctx context.Context, page int) ([]dto.LoanData, error)
	
	Confirm(ctx context.Context, data dto.Confirm) error
	
	AddCategory(ctx context.Context, data dto.Category) error
	AddBook(ctx context.Context, data dto.BookData) error
}
