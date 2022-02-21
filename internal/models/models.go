package models

import (
	"time"
)

// User is the user model
type Users struct {
	ID              int
	FirstName       string
	LastName        string
	Email           string
	IsEmailVerified bool
	Phone           string
	Password        string
	AccessLevel     int
	Token           string
	TokenExpiry     time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Property struct {
	ID               int
	PropertyName     string
	PropertyLocation string
	OwnerID          int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Unit struct {
	ID              int
	UnitName        string
	PropertyID      int
	OccupancyStatus bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Invoice struct {
	ID             int
	InvoiceName    string
	UnitID         int
	MonthlyAmount  int
	AmountReceived int
	AmountDue      int
	DatePaid       time.Time
	DateDue        time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Tenant struct {
	ID                          int
	FirstName                   string
	LastName                    string
	Email                       string
	Phone                       string
	OtherPhone                  string
	AlternateContactPersonName  string
	AlternateContactPersonPhone string
	RiskID                      int
	UnitID                      int
	DateOfOccupancy             time.Time
	ExitDate                    time.Time
	InvoiceID                   int
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
}

type Expenses struct {
	ID          int
	UnitID      int
	ExpenseName string
	DatePaid    time.Time
	Narration   string
	AmountPaid  int
	AmountDue   int
	DueDate     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Risk struct {
	ID        int
	RiskName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// MailData holds an email message
type MailData struct {
	To       string
	From     string
	Subject  string
	Content  string
	Template string
}
