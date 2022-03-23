package dbrepo

import (
	"github.com/Gaussgeek/rentals/internal/models"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

func (m *testDBRepo) GetUserByID(id int) (models.Users, error) {
	var u models.Users

	return u, nil
}

func (m *testDBRepo) UpdateUser(u models.Users) error {
	return nil
}

func (m *testDBRepo) InsertNewUser(r models.Users) error {
	return nil
}

//AddNewProperty adds a new property to the database
func (m *testDBRepo) AddNewProperty(p models.Property) error {
	return nil
}

func (m *testDBRepo) GetPropertiesByOwnwerID(id int) ([]models.Property, error) {
	var properties []models.Property

	return properties, nil
}

func (m *testDBRepo) GetPropertyByPropertyID(id int) (models.Property, error) {
	var property models.Property

	return property, nil
}

func (m *testDBRepo) InsertNewUnit(models.Unit) error {
	return nil
}

func (m *testDBRepo) UpdateUnitDetails(u models.Unit) error {
	return nil
}

func (m *testDBRepo) GetUnitsByPropertyID(id int) ([]models.Unit, error) {
	var units []models.Unit

	return units, nil
}

func (m *testDBRepo) GetUnitByUnitID(id int) (models.Unit, error) {
	var unit models.Unit
	return unit, nil
}
func (m *testDBRepo)GetTenantByUnitID(id int) (models.Tenant, error) {
	var t models.Tenant
	return t, nil
}

func (m *testDBRepo) GetInvoiceByUnitID(id int) (models.Invoice, error) {
	var i models.Invoice
	return i, nil
}

func (m *testDBRepo) GetExpenseByUnitID(id int) (models.Expenses, error) {
	var exp models.Expenses
	return exp, nil
}

func (m *testDBRepo)InsertNewTenant(u models.Tenant) error {
		return nil
}

func (m *testDBRepo)UpdateTenant(u models.Tenant) error {
	return nil
}

func (m *testDBRepo) DeleteTenant(id int) error {
	return nil
}

func (m *testDBRepo) InsertNewExpense(u models.Expenses) error {
	return nil
}

func (m *testDBRepo) UpdateExpense(u models.Expenses) error {
	return nil
}

func (m *testDBRepo) InsertNewInvoice(u models.Invoice) error {
	return nil
}

func (m *testDBRepo)GetInvoicesByUnitID(id int) ([]models.Invoice, error) {
	var invoices []models.Invoice
	return invoices, nil
}

func (m *testDBRepo) GetInvoiceByInvoiceID(id int) (models.Invoice, error) {
	var u models.Invoice
	return u, nil
}

func (m *testDBRepo) UpdateInvoice(u models.Invoice) error {
	return nil
}

func (m *testDBRepo) GetAllUnitsByOwnerID(id int) ([]models.Unit, error) {
	var u []models.Unit
	return u, nil
}

func (m *testDBRepo) AddNewTokenToUser(id int, s string) error {
	return nil
}

func (m *testDBRepo)SetEmailVerifiedTrue(id int) error {
	return nil
}

func (m *testDBRepo)GetAllTenantsByOwnerID(id int) ([]models.Tenant, error) {
	var t []models.Tenant
	return t, nil
}

func (m *testDBRepo) GetOverDueExpenses(id int) ([]models.Expenses, error) {
	var expenses []models.Expenses
	return expenses, nil
}

func (m *testDBRepo) GetOverDueInvoices(id int) ([]models.Invoice, error) {
	var invoices []models.Invoice
	return invoices, nil

}