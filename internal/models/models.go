package models

import (
	"time"
)

// User is the user model
type User struct {
	ID          int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	AccessLevel int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// MailData holds an email message
type MailData struct {
	To       string
	From     string
	Subject  string
	Content  string
	Template string
}
