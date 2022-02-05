package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Gaussgeek/rentals/internal/config"
	"github.com/Gaussgeek/rentals/internal/driver"
	"github.com/Gaussgeek/rentals/internal/forms"
	"github.com/Gaussgeek/rentals/internal/helpers"
	"github.com/Gaussgeek/rentals/internal/models"
	"github.com/Gaussgeek/rentals/internal/render"
	"github.com/Gaussgeek/rentals/internal/repository"
	"github.com/Gaussgeek/rentals/internal/repository/dbrepo"
	"golang.org/x/crypto/bcrypt"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewTestRepo creates a new repository
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingsRepo(a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

//Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// send the data to the template
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Contact renders the search availability page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// SignUp renders the signup page
func (m *Repository) SignUp(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "signup.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// HashPassword returns password in hashed form
func HashPassword(password string) (string, error) {
	pw := []byte(password)

	hashedPassword, err := bcrypt.GenerateFromPassword(pw, 12)

	if err != nil {
		panic(err)
	}

	return string(hashedPassword), err
}

// PostSignUp renders the signup page
func (m *Repository) PostSignUp(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	first_name := r.Form.Get("first_name")
	last_name := r.Form.Get("last_name")
	email := r.Form.Get("email")
	password, _ := HashPassword(r.Form.Get("password"))

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		render.Template(w, r, "signup.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	NewUser := models.Users{
		FirstName: first_name,
		LastName:  last_name,
		Email:     email,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = m.DB.InsertNewUser(NewUser)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert new user into database!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "NewUser", NewUser)
	m.App.Session.Put(r.Context(), "flash", "Successfully added to the database.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// ShowLogin shows the login screen
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostShowLogin handles logging the user in
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout logs a user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

//AdminAddNewProperty renders the add new property page
func (m *Repository) AdminAddNewProperty(w http.ResponseWriter, r *http.Request) {
	user_id, ok := m.App.Session.Get(r.Context(), "user_id").(int)

	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get your id from session")
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	data := make(map[string]int)
	data["user_id"] = user_id

	render.Template(w, r, "admin.add-new-property.page.tmpl", &models.TemplateData{
		Form:   forms.New(nil),
		IntMap: data,
	})
}

//AdminPostAddNewProperty is the handler for the PostAddNew property
func (m *Repository) AdminPostAddNewProperty(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())
	owner_id, ok := m.App.Session.Get(r.Context(), "user_id").(int)

	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get your id from session")
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	form := forms.New(r.PostForm)
	form.Required("property_name", "property_location")

	if !form.Valid() {
		render.Template(w, r, "admin.add-new-property.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	NewProperty := models.Property{
		PropertyName:     r.Form.Get("property_name"),
		PropertyLocation: r.Form.Get("property_location"),
		OwnerID:          owner_id,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err = m.DB.AddNewProperty(NewProperty)

	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert new property into database!")
		http.Redirect(w, r, "/admin/new-property", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "NewProperty", NewProperty)
	m.App.Session.Put(r.Context(), "flash", "Successfully added new property to the database.")

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

//AdminAllPropertiesByID handles the display of all properties for an owner id
func (m *Repository) AdminAllPropertiesByID(w http.ResponseWriter, r *http.Request) {
	id, ok := m.App.Session.Get(r.Context(), "user_id").(int)

	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get your id from session")
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	properties, err := m.DB.GetPropertiesByOwnwerID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["properties"] = properties

	render.Template(w, r, "admin-all-properties.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminShowPropertyByPropertyID handles display of a property
func (m *Repository) AdminShowPropertyByPropertyID(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[2]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	property, err := m.DB.GetPropertyByPropertyID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["property"] = property

	render.Template(w, r, "admin-show-property.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

// AdminAddUnitToProperty handles the page for adding a unit to a property
func (m *Repository) AdminAddUnitToProperty(w http.ResponseWriter, r *http.Request) {
	owner_id, ok := m.App.Session.Get(r.Context(), "user_id").(int)

	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get your id from session")
		http.Redirect(w, r, fmt.Sprintf("/admin/all-properties/%d/show", owner_id), http.StatusSeeOther)
		return
	}

	exploded := strings.Split(r.RequestURI, "/")

	property_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	property, err := m.DB.GetPropertyByPropertyID(property_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["property"] = property

	render.Template(w, r, "admin-add-unit-to-property.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminPostAddUnitToProperty handles the post for adding unit to property
func (m *Repository) AdminPostAddUnitToProperty(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	form := forms.New(r.PostForm)
	form.Required("unit_name")

	if !form.Valid() {
		render.Template(w, r, "admin.dashboard.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	property_id, err := strconv.Atoi(r.Form.Get("property_id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	NewUnit := models.Unit{
		UnitName:        r.Form.Get("unit_name"),
		PropertyID:      property_id,
		OccupancyStatus: false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err = m.DB.InsertNewUnit(NewUnit)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Successfully added new unit to the database.")

	http.Redirect(w, r, fmt.Sprintf("/admin/all-properties/%d/show", property_id), http.StatusSeeOther)

}

// AdminShowUnitsByPropertyID handles display of all units at a property
func (m *Repository) AdminShowUnitsByPropertyID(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	property_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	units, err := m.DB.GetUnitsByPropertyID(property_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["units"] = units

	render.Template(w, r, "admin-all-units-on-property.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminShowUnitDetails handles display of a unit page
func (m *Repository) AdminShowUnitDetails(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	unit_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	unit, err := m.DB.GetUnitByUnitID(unit_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["unit"] = unit

	render.Template(w, r, "admin-unit-details.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminAddTenantByUnitID adds a new tenant to a unit
func (m *Repository) AdminAddTenantByUnitID(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	unit_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	
	intmap := make(map[string]int)
	intmap["unit_id"] = unit_id
	
	render.Template(w, r, "unit-add-new-tenant.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		IntMap: intmap,
	})
}

// AdminPostAddTenantByUnitID is the post handler for adding tenant to unit
func (m *Repository) AdminPostAddTenantByUnitID(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "phone", "email", "date_of_occupancy", "unit_id")

	if !form.Valid() {
		render.Template(w, r, "unit-add-new-tenant.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	unit_id, err := strconv.Atoi(r.Form.Get("unit_id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}


	sd := r.Form.Get("date_of_occupancy")
	ed := r.Form.Get("exit_date")

	layout := "2006-01-02"

	date_of_occupancy, err := time.Parse(layout, sd)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	exit_date, err := time.Parse(layout, ed)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse end date")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}


	NewTenant := models.Tenant{
		FirstName:  r.Form.Get("first_name"),
		LastName: r.Form.Get("last_name"),
		Email: r.Form.Get("email"),
		Phone: r.Form.Get("phone"),
		OtherPhone: r.Form.Get("other_phone"),
		AlternateContactPersonName: r.Form.Get("other_contact_name"),
		AlternateContactPersonPhone: r.Form.Get("other_contact_phone"),
		RiskID: 1,
		UnitID: unit_id,
		DateOfOccupancy: date_of_occupancy,
		ExitDate: exit_date,
		InvoiceID: 1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	fmt.Println(NewTenant)

	err = m.DB.InsertNewTenant(NewTenant)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert tenant into database!")
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Successfully added new tenant to the database.")

	http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
}

// AdminShowTenantDetails displays tenant details
func (m *Repository) AdminShowTenantDetails(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	unit_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	
	intmap := make(map[string]int)
	intmap["unit_id"] = unit_id

	tenant, err := m.DB.GetTenantByUnitID(unit_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["tenant"] = tenant
	
	render.Template(w, r, "admin-view-tenant.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		IntMap: intmap,
		Data: data,
	})
}

// AdminDeleteTenant deletes a tenant from a database
func (m *Repository) AdminDeleteTenant(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	unit_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	tenant_id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = m.DB.DeleteTenant(tenant_id)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't delete tenant from database!")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/view-tenants", unit_id), http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "success", "Deleted Tenant from database.")

	http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
}

// AdminUpdateTenantByID updates a tenant by tenant ID
func (m *Repository) AdminUpdateTenantByID(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "phone", "email", "date_of_occupancy", "unit_id")

	if !form.Valid() {
		render.Template(w, r, "admin-view-tenant.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}
	
	exploded := strings.Split(r.RequestURI, "/")

	unit_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	tenant_id, err := strconv.Atoi(exploded[5])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	sd := r.Form.Get("date_of_occupancy")
	ed := r.Form.Get("exit_date")
	cAt := r.Form.Get("created_at")

	layout := "2006-01-02"

	date_of_occupancy, err := time.Parse(layout, sd)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	exit_date, err := time.Parse(layout, ed)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse end date")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	created_at, err := time.Parse(layout, cAt)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't created at date")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	tenant, err := m.DB.GetTenantByUnitID(unit_id)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't find tenant")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	tenant.ID = tenant_id
	tenant.FirstName =  r.Form.Get("first_name")
	tenant.LastName = r.Form.Get("last_name")
	tenant.Email = r.Form.Get("email")
	tenant.Phone = r.Form.Get("phone")
	tenant.OtherPhone = r.Form.Get("other_phone")
	tenant.AlternateContactPersonName = r.Form.Get("other_contact_name")
	tenant.AlternateContactPersonPhone = r.Form.Get("other_contact_phone")
	tenant.RiskID = 1
	tenant.UnitID = unit_id
	tenant.DateOfOccupancy = date_of_occupancy
	tenant.ExitDate = exit_date
	tenant.InvoiceID = 1
	tenant.CreatedAt = created_at
	tenant.UpdatedAt = time.Now()

	err = m.DB.UpdateTenant(tenant)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Updated tenant details")
	
	http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)

}