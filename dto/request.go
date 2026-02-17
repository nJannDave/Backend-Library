package dto

type Students struct {
	NIS         int    `json:"nis" binding:"required,number"`
	Name        string `json:"name" binding:"required,min=5,max=30"`
	PhoneNumber string `json:"phone_number" binding:"required,max=15"`
	Email       string `json:"email" binding:"required,email,min=10,max=20"`
	Password    string `json:"password" binding:"required"`
	Class       string `json:"class" binding:"required`
	SubClass    string `json:"sub_class" binding:"required,oneof=A B C"`
	Major       string `json:"major" binding:"required,oneof=RPL SIJA PSPT TPTU TEI MEKA TOI TEK IOP"`
	Batch       int    `json:"batch" binding:"required,number"`
}

type Login struct {
	NIS      int    `json:"nis" binding:"number,required"`
	Password string `json:"password" binding:"required"`
}

type Loan struct {
	ID         int    `json:"book_id" binding:"required,number"`
	ReturnedAt string `json:"returned_at" binding:"required"`
}

type Category struct {
	Name string `json:"category_name" binding:"required"`
}

type BookData struct {
	ISBN           string `json:"isbn" binding:"required"`
	Name           string `json:"book_name" binding:"required"`
	Author         string `json:"author" binding:"required"`
	Publisher      string `json:"publisher" binding:"required"`
	Description    string `json:"description" binding:"required,min=50,max=300"`
	Stock          int    `json:"stock" binding:"required,number"`
	AvailableStock int    `json:"available_stock" binding:"required,number"`
	IDCategory     []int  `json:"id_category" binding:"required,dive,gt=0"`
}

type Confirm struct {
	Student string `json:"student" binding:"required"`
	ISBN string `json:"isbn" binding:"required"`
}