package repository

import (
	"github.com/Gaussgeek/rentals/internal/models"
)

// DatabaseRepo is the list of all functions working on the database
type DatabaseRepo interface {
	AllUsers() bool

	InsertNewUser(r models.Users) error

	AddNewTokenToUser(id int, s string) error
	SetEmailVerifiedTrue(id int) error

	GetUserByID(id int) (models.Users, error)
	UpdateUser(u models.Users) error
	Authenticate(email, testPassword string) (int, string, error)

	AddNewProperty(p models.Property) error
	GetPropertiesByOwnwerID(id int) ([]models.Property, error)

	GetPropertyByPropertyID(id int) (models.Property, error)

	InsertNewUnit(models.Unit) error
	UpdateUnitDetails(u models.Unit) error

	InsertNewTenant(u models.Tenant) error
	UpdateTenant(u models.Tenant) error
	GetAllTenantsByOwnerID(id int) ([]models.Tenant, error)
	DeleteTenant(id int) error

	GetUnitsByPropertyID(id int) ([]models.Unit, error)
	GetAllUnitsByOwnerID(id int) ([]models.Unit, error)

	GetUnitByUnitID(id int) (models.Unit, error)
	GetTenantByUnitID(id int) (models.Tenant, error)
	GetInvoiceByUnitID(id int) (models.Invoice, error)
	GetExpenseByUnitID(id int) (models.Expenses, error)

	InsertNewExpense(u models.Expenses) error
	UpdateExpense(u models.Expenses) error
	GetOverDueExpenses(id int) ([]models.Expenses, error)

	InsertNewInvoice(u models.Invoice) error
	GetInvoicesByUnitID(id int) ([]models.Invoice, error)
	GetOverDueInvoices(id int) ([]models.Invoice, error)

	GetInvoiceByInvoiceID(id int) (models.Invoice, error)
	UpdateInvoice(u models.Invoice) error 
}
