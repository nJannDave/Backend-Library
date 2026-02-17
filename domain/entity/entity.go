package entity

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Students struct {
	NIS          int
	PersonalInfo PersonalInfo `gorm:"embedded"`
	Password     string
	AcademicInfo AcademicInfo `gorm:"embedded"`
}

type PersonalInfo struct {
	Name        string
	PhoneNumber string
	Email       string
}

type AcademicInfo struct {
	Class    string
	SubClass string
	Major    string
	Batch    int
}

var phoneRegex = regexp.MustCompile(`^\+?[0-9]+$`)

func (p *PersonalInfo) ValidatePN(errMsg *[]string) {
	cleanNumber := strings.NewReplacer(
		"-", "",
		"(", "",
		")", "",
		".", "",
		" ", "",
	).Replace(p.PhoneNumber)
	p.PhoneNumber = cleanNumber
	if len(cleanNumber) < 8 || len(cleanNumber) > 15 {
		*errMsg = append(*errMsg, "number length must be 8-15")
		return
	}
	if !phoneRegex.MatchString(cleanNumber) {
		*errMsg = append(*errMsg, "only numeric characters are allowed")
		return
	}
	var realNumber string
	if strings.HasPrefix(cleanNumber, "+62") {
		realNumber = cleanNumber
	} else if strings.HasPrefix(cleanNumber, "62") {
		realNumber = "+" + cleanNumber
	} else if strings.HasPrefix(cleanNumber, "0") {
		realNumber = "+62" + strings.TrimPrefix(cleanNumber, "0")
	} else {
		*errMsg = append(*errMsg, "prefix format must start with 0 / 62 / +62")
		return
	}
	p.PhoneNumber = realNumber
}

func (a *AcademicInfo) ValidateClass(errMsg *[]string) {
	isIOPorSIJA := a.Major == "IOP" || a.Major == "SIJA"

	if isIOPorSIJA {
		if a.Class != "X" && a.Class != "XI" && a.Class != "XII" && a.Class != "XIII" {
			*errMsg = append(*errMsg, "class not available for IOP/SIJA (max XIII)")
		}
	} else {
		if a.Class != "X" && a.Class != "XI" && a.Class != "XII" {
			*errMsg = append(*errMsg, "class not available (max XII)")
		}
	}}

func (Students) TableName() string {
	return "students"
}

type Login struct {
	NIS      int
	Password string
}

func (Login) TableName() string {
	return "students"
}

type Categories struct {
	ID   int `gorm:"primaryKey"`
	Name string
}

func (Categories) TableName() string {
	return "categories"
}

type Category struct {
	Name string
}

func (Category) TableName() string {
	return "categories"
}

type Book struct {
	ID             int `gorm:"primaryKey"`
	Name           string
	Author         string
	Publisher      string
	Description    string
	Categories     []Categories `gorm:"many2many:connections;joinForeignKey:id_book;joinReferences:id_category"`
	AvailableStock int
}

func (Book) TableName() string {
	return "books"
}

type Loan struct {
	IdUser         int
	IdBook         int
	MustReturnedAt time.Time
}

func (l *Loan) ValidateDateFormat(date string) error {
	const layout = "02-01-2006"
	formattedTime, err := time.Parse(layout, date)
	if err != nil {
		return fmt.Errorf("wrong format make sure it is like this: dd-mm-yyyy")
	}
	l.MustReturnedAt = formattedTime
	return nil
}

func (l *Loan) ValidateDate() error {
	limit := time.Now().AddDate(0, 0, 7)
	if l.MustReturnedAt.After(limit) {
		return fmt.Errorf("maximum loan limit is 7 days")
	}
	if l.MustReturnedAt.Before(time.Now()) {
		return fmt.Errorf("date cannot be in the past")
	}
	return nil
}

func (Loan) TableName() string {
	return "loan"
}

type Confirm struct {
	Student string
	ISBN string
}

type LoanData struct {
	BookName       string     `gorm:"column:book_name"`
	StudentName    string     `gorm:"column:student_name"`
	BorrowAt       time.Time  `gorm:"column:borrow_at"`
	MustReturnedAt time.Time  `gorm:"column:must_returned_at"`
	ReturnedAt     *time.Time `gorm:"column:returned_at"`
	Sanctions      *int64     `gorm:"column:sanctions"`
}

func (LoanData) TableName() string {
	return "loan"
}

type LdUpdate struct {
	MustReturnedAt time.Time `gorm:"column:must_returned_at"`
	ReturnedAt *time.Time `gorm:"column:returned_at"`
	Sanctions *int64 `gorm:"column:sanctions"`
}

func (ldu *LdUpdate) GiveSanctions() {
	if ldu.ReturnedAt.After(ldu.MustReturnedAt) {
		d := (int64(time.Since(ldu.MustReturnedAt).Hours()) / 24)
		*ldu.Sanctions = d * 2000
	}
	if ldu.ReturnedAt.Before(ldu.MustReturnedAt) {
		*ldu.Sanctions = 0
	}
}

func (LdUpdate) TableName() string {
	return "loan"
}

type BookData struct {
	BookID         int `gorm:"column:id"`
	ISBN           string
	Name           string
	Author         string
	Publisher      string
	Description    string
	Stock          int
	AvailableStock int
}

func (BookData) TableName() string {
	return "books"
}

type Connections struct {
	BookID     int `gorm:"column:id_book"`
	IdCategory int `gorn:"column:id_category"`
}

func (Connections) TableName() string {
	return "connections"
}
