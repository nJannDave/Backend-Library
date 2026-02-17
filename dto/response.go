package dto

import "time"

type Binding struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type Service struct {
	Reason string `json:"reason"`
}

type Errors struct {
	Binding []Binding `json:"binding,omitzero"`
	Service []Service `json:"service,omitzero"`
	Error   string    `json:"error,omitzero"`
}

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitzero"`
	Error   Errors      `json:"errors,omitzero"`
	Data    interface{} `json:"data,omitzero"`
}

type Categories struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Books struct {
	ID             int          `json:"book_id" redis:"id"`
	Name           string       `json:"name" redis:"name"`
	Author         string       `json:"author" redis:"author"`
	Publisher      string       `json:"publisher" redis:"publisher"`
	Description    string       `json:"description" redis:"description"`
	Categories     []Categories `json:"categories" redis:"categories"`
	AvailableStock int          `json:"available_stock" redis:"available_stock"`
}

type LoanData struct {
	BookName       string     `json:"book_name"`
	StudentName    string     `json:"student_name"`
	BorrowAt       time.Time  `json:"borrow_at"`
	MustReturnedAt time.Time  `json:"must_returned_at"`
	ReturnedAt     *time.Time `json:"returned_at"`
	Sanctions      *int64     `json:"sanctions"`
}
