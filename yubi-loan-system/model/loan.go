package model

import "github.com/jinzhu/gorm"

type Loan struct {
	gorm.Model
	PersonName    string
	Age           int
	LoanAmount    float64
	PanCardNumber string
	AnnualIncome  float64
	DocURL        string
	S3URL         string
	Status        string
}
