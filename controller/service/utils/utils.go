package utils

import (
	"encoding/json"
	"stmnplibrary/domain/entity"
	"stmnplibrary/dto"
	"time"

	"fmt"
	"strings"
)

func ValidateErr(err error, subStr string, errMsg *[]string) error {
	if strings.Contains(err.Error(), subStr) {
		*errMsg = append(*errMsg, err.Error())
		return nil
	}
	return err
}

func ValidateErrTw(err error, errMsg string) error {
	if strings.Contains(err.Error(), "internal server error: ") || strings.Contains(err.Error(), "operation" ) {
		return fmt.Errorf(errMsg, err)
	}
	return err
}

func ValidateErrLoan(err error, opt string) error {
	if strings.Contains(err.Error(), "no data affected") && opt == "user" {
		return fmt.Errorf("has reached the limit")
	}
	if strings.Contains(err.Error(), "no data affected") && opt == "book" {
		return fmt.Errorf("out of stock")
	}
	return ValidateErrTw(err, "service - loan: %w")
}

func BooksMapper(result []entity.Book) []dto.Books {
	var books []dto.Books
	for _, v := range result {
		var categories []dto.Categories
		for _, c := range v.Categories {
			categories = append(categories, dto.Categories{
				ID:   c.ID,
				Name: c.Name,
			})
		}
		books = append(books, dto.Books{
			ID:             v.ID,
			Name:           v.Name,
			Author:         v.Author,
			Publisher:      v.Publisher,
			Description:    v.Description,
			Categories:     categories,
			AvailableStock: v.AvailableStock,
		})
	}
	return books
}

func LoanDataMapper(ld []entity.LoanData) []dto.LoanData {
	var loanData = make([]dto.LoanData, 0, 35)
	for _, i := range ld {
		var b dto.LoanData
		b.BookName = i.BookName
		b.BorrowAt = i.BorrowAt
		b.MustReturnedAt = i.MustReturnedAt
		b.ReturnedAt = i.ReturnedAt
		b.Sanctions = i.Sanctions
		b.StudentName = i.StudentName
		loanData = append(loanData, b)
	}
	return loanData
}

func UnMarshal(dataByte []byte, data any) error {
	if err := json.Unmarshal(dataByte, data); err != nil {
		return fmt.Errorf("failed unmarshal: %w", err)
	}
	return nil
}

func Marshal(data any) ([]byte, error) {
	val, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed marshal: %w", err)
	}
	return val, nil
}

func InitLD(data *entity.LdUpdate) entity.LdUpdate {
	var (
		rA = time.Now()
		sT int64 = 0
	)
	return entity.LdUpdate{
		MustReturnedAt: data.MustReturnedAt,
		Sanctions: &sT,
		ReturnedAt: &rA,
	}
}