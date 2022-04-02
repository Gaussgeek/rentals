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
	"github.com/dchest/uniuri"
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

/*
func (m *Repository) randomString(n int) string {
	return h.App.RandomString(n)
}

func (m *Repository) encrypt(text string) (string, error) {
	enc := Encryption{Key: []byte(h.App.EncryptionKey)}

	encrypted, err := enc.Encrypt(text)
	if err != nil {
		return "", err
	}
	return encrypted, nil
}

func (m *Repository) decrypt(crypto string) (string, error) {
	enc := Encryption{Key: []byte(h.App.EncryptionKey)}

	decrypted, err := enc.Decrypt(crypto)
	if err != nil {
		return "", err
	}
	return decrypted, nil
}
*/

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

	user, err := m.DB.GetUserByID(id)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't get user details from database")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	
	if !user.IsEmailVerified {
		m.App.Session.Put(r.Context(), "error", "Verify your email to continue")
		http.Redirect(w, r, fmt.Sprintf("/user/verify-email/%d", id), http.StatusSeeOther)
		return
	}

	
	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// VerifyUserEmail handles verification of user email
func (m *Repository) VerifyUserEmail(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	user, err := m.DB.GetUserByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	token := uniuri.NewLen(128)

	err = m.DB.AddNewTokenToUser(id, token)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// TO DO: encrypt the token before sending to user email

	// send email to user
	htmlMessage := fmt.Sprintf(
					`<strong>Email Verification</strong><br>
					Dear %s: <br>
					Click the link below to verify your email inorder to continue
					using Property Manager App <br><br><br><br>

					<a href="localhost:8080/user/verify-link/%d/%s"> localhost:8080/user/verify-link/%d/%s</a>
					<br><br><br>
					In case of a delayed response, copy and paste the link in your browser.
					` , user.FirstName, user.ID, token, user.ID, token)
	
	msg := models.MailData{
		To: user.Email,
		From: "info@mulisa.com",
		Subject: "Email Verification",
		Content: htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	render.Template(w, r, "email-verify-send-notice.page.tmpl", &models.TemplateData{})
}

// VerifyEmail handles verification of token
func (m *Repository) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	token := exploded[4]
	
	user, err := m.DB.GetUserByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if time.Now().After(user.TokenExpiry) {
		m.App.Session.Put(r.Context(), "error", "This token has expired. A new one has been sent to your email.")
		http.Redirect(w, r, fmt.Sprintf("/user/verify-email/%d", id), http.StatusSeeOther)
		return
	}

	if !(token == user.Token) {
		m.App.Session.Put(r.Context(), "error", "Invalid verification token.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err = m.DB.SetEmailVerifiedTrue(id)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Unable to verify your email")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Email has been verified successfully.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)

}

// Logout logs a user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	user_id, ok := m.App.Session.Get(r.Context(), "user_id").(int)

	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get your id from session")
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["user_id"] = user_id
	
	tenants, err := m.DB.GetAllTenantsByOwnerID(user_id)
	if err != nil {
		helpers.ServerError(w, err)
	}

	properties, err := m.DB.GetPropertiesByOwnwerID(user_id)
	if err != nil {
		helpers.ServerError(w, err)
	}

	units, err := m.DB.GetAllUnitsByOwnerID(user_id)
	if err != nil {
		helpers.ServerError(w, err)
	}

	dueInvoices, err := m.DB.GetOverDueInvoices(user_id)
	if err != nil {
		helpers.ServerError(w, err)
	}

	dueExpenses, err := m.DB.GetOverDueExpenses(user_id)
	if err != nil {
		helpers.ServerError(w, err)
	}

	data["tenants"] = tenants
	data["properties"] = properties
	data["units"] = units
	data["dueInvoices"] = dueInvoices
	data["dueExpenses"] = dueExpenses

	intmap := make(map[string]int)
	intmap["NumTenants"] = len(tenants)
	intmap["NumProperties"] = len(properties)
	intmap["NumUnits"] = len(units)

	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{
		Data: data,
		IntMap: intmap,
	})
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

// AdminUpdateUnit updates a unit name
func (m *Repository) AdminUpdateUnit(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	exploded := strings.Split(r.RequestURI, "/")

	unit_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	
	unit, err := m.DB.GetUnitByUnitID(unit_id)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't find unit from database")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	unit.OccupancyStatus, err = strconv.ParseBool(r.Form.Get("occupancy_status"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	unit.UnitName = r.Form.Get("unit_name")
	unit.UpdatedAt = time.Now()

	prop_id := unit.PropertyID

	err = m.DB.UpdateUnitDetails(unit)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Updated unit details")

	http.Redirect(w, r, fmt.Sprintf("/admin/all-properties/%d/view-units", prop_id), http.StatusSeeOther)

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
		Form:   forms.New(nil),
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
		FirstName:                   r.Form.Get("first_name"),
		LastName:                    r.Form.Get("last_name"),
		Email:                       r.Form.Get("email"),
		Phone:                       r.Form.Get("phone"),
		OtherPhone:                  r.Form.Get("other_phone"),
		AlternateContactPersonName:  r.Form.Get("other_contact_name"),
		AlternateContactPersonPhone: r.Form.Get("other_contact_phone"),
		RiskID:                      1,
		UnitID:                      unit_id,
		DateOfOccupancy:             date_of_occupancy,
		ExitDate:                    exit_date,
		InvoiceID:                   1,
		CreatedAt:                   time.Now(),
		UpdatedAt:                   time.Now(),
	}

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

	if tenant.FirstName == "" {
		m.App.Session.Put(r.Context(), "error", "There are no tenants yet. Add a new tenant")
		http.Redirect(w, r, fmt.Sprintf("/admin//unit-details/%d/add-new-tenant", unit_id), http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["tenant"] = tenant

	render.Template(w, r, "admin-view-tenant.page.tmpl", &models.TemplateData{
		Form:   forms.New(nil),
		IntMap: intmap,
		Data:   data,
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

	tenant, err := m.DB.GetTenantByUnitID(unit_id)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't find tenant")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	tenant.ID = tenant_id
	tenant.FirstName = r.Form.Get("first_name")
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
	tenant.UpdatedAt = time.Now()

	err = m.DB.UpdateTenant(tenant)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Updated tenant details")

	http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)

}

// AdminAddNewExpense adds a new expense
func (m *Repository) AdminAddNewExpense(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	unit_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	intmap := make(map[string]int)
	intmap["unit_id"] = unit_id

	render.Template(w, r, "admin-add-new-expense.page.tmpl", &models.TemplateData{
		Form:   forms.New(nil),
		IntMap: intmap,
	})
}

// AdminPostAddNewExpense handles adding new expense
func (m *Repository) AdminPostAddNewExpense(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	form := forms.New(r.PostForm)
	form.Required("expense_name", "amount_paid", "narration", "date_paid", "due_date", "unit_id")

	if !form.Valid() {
		render.Template(w, r, "admin-add-new-expense.page.tmpl", &models.TemplateData{
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

	amount_paid, err := strconv.Atoi(r.Form.Get("amount_paid"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	amount_due, err := strconv.Atoi(r.Form.Get("amount_due"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	d_paid := r.Form.Get("date_paid")
	du_date := r.Form.Get("due_date")

	layout := "2006-01-02"

	date_paid, err := time.Parse(layout, d_paid)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse paid date")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	due_date, err := time.Parse(layout, du_date)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse end date")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	NewExpense := models.Expenses{
		ExpenseName: r.Form.Get("expense_name"),
		UnitID:      unit_id,
		AmountPaid:  amount_paid,
		DatePaid:    date_paid,
		Narration:   r.Form.Get("narration"),
		AmountDue:   amount_due,
		DueDate:     due_date,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = m.DB.InsertNewExpense(NewExpense)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert expense into database!")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/add-new-expenses", unit_id), http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Successfully added new expense to the database.")

	http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
}

// AdminGetExpensesByUnitID displays unit expenses
func (m *Repository) AdminGetExpensesByUnitID(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	unit_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	intmap := make(map[string]int)
	intmap["unit_id"] = unit_id

	expense, err := m.DB.GetExpenseByUnitID(unit_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if expense.ExpenseName == "" {
		m.App.Session.Put(r.Context(), "error", "There are no Expenses yet. Add a new expense")
		http.Redirect(w, r, fmt.Sprintf("/admin//unit-details/%d/add-new-expenses", unit_id), http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["expense"] = expense

	render.Template(w, r, "admin-view-expenses.page.tmpl", &models.TemplateData{
		Form:   forms.New(nil),
		IntMap: intmap,
		Data:   data,
	})
}

// AdminUpdateExpenseByID updates an expense
func (m *Repository) AdminUpdateExpenseByID(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	exploded := strings.Split(r.RequestURI, "/")

	unit_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	
	amount_paid, err := strconv.Atoi(r.Form.Get("amount_paid"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	amount_due, err := strconv.Atoi(r.Form.Get("amount_due"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	d_paid := r.Form.Get("date_paid")
	du_date := r.Form.Get("due_date")

	layout := "2006-01-02"

	date_paid, err := time.Parse(layout, d_paid)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse paid date")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	due_date, err := time.Parse(layout, du_date)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse end date")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	UnitExpense, err := m.DB.GetExpenseByUnitID(unit_id)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't find tenant")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	UnitExpense.ExpenseName = r.Form.Get("expense_name")
	UnitExpense.AmountPaid = amount_paid
	UnitExpense.DatePaid = date_paid
	UnitExpense.AmountDue = amount_due
	UnitExpense.DueDate = due_date
	UnitExpense.Narration = r.Form.Get("narration")
	UnitExpense.UpdatedAt = time.Now()

	err = m.DB.UpdateExpense(UnitExpense)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Updated Expense details")

	http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
}

// AdminAddNewInvoice adds a new invoice
func (m *Repository) AdminAddNewInvoice(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	unit_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	intmap := make(map[string]int)
	intmap["unit_id"] = unit_id

	render.Template(w, r, "admin-add-new-invoice.page.tmpl", &models.TemplateData{
		Form:   forms.New(nil),
		IntMap: intmap,
	})
}

//  AdminPostAddNewInvoice handles post add of a new invoice
func (m *Repository) AdminPostAddNewInvoice(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	form := forms.New(r.PostForm)
	form.Required("invoice_name", "amount_paid", "date_paid", "due_date", "unit_id")

	if !form.Valid() {
		render.Template(w, r, "admin-add-new-invoice.page.tmpl", &models.TemplateData{
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

	monthly_amount, err := strconv.Atoi(r.Form.Get("monthly_amount"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	amount_paid, err := strconv.Atoi(r.Form.Get("amount_paid"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	amount_due, err := strconv.Atoi(r.Form.Get("amount_due"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	d_paid := r.Form.Get("date_paid")
	du_date := r.Form.Get("due_date")

	layout := "2006-01-02"

	date_paid, err := time.Parse(layout, d_paid)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse paid date")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	due_date, err := time.Parse(layout, du_date)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse due date")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	NewInvoice := models.Invoice{
		InvoiceName: r.Form.Get("invoice_name"),
		UnitID:      unit_id,
		MonthlyAmount: monthly_amount,
		AmountReceived:  amount_paid,
		DatePaid:    date_paid,
		AmountDue:   amount_due,
		DateDue:     due_date,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = m.DB.InsertNewInvoice(NewInvoice)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert expense into database!")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/add-new-expenses", unit_id), http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Successfully added new invoice to the database.")

	http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
}

// AdminShowInvoices displays invoices by unit id
func (m *Repository) AdminShowInvoicesByUnitID(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	unit_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	intmap := make(map[string]int)
	intmap["unit_id"] = unit_id

	invoices, err := m.DB.GetInvoicesByUnitID(unit_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["invoices"] = invoices

	render.Template(w, r, "admin-all-invoices.page.tmpl", &models.TemplateData{
		Data: data,
		IntMap: intmap,
	})
}

//  AdminShowInvoiceByInvoiceID dosplays invoice details
func (m *Repository) AdminShowInvoiceByInvoiceID(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	invoice_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	invoice, err := m.DB.GetInvoiceByInvoiceID(invoice_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	intmap := make(map[string]int)
	intmap["unit_id"] = invoice.UnitID

	data := make(map[string]interface{})
	data["invoice"] = invoice

	render.Template(w, r, "admin-one-invoice.page.tmpl", &models.TemplateData{
		Data: data,
		IntMap: intmap,
	})
}

// AdminUpdateInvoiceByInvoiceID updates an invoice
func (m *Repository) AdminUpdateInvoiceByInvoiceID(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	exploded := strings.Split(r.RequestURI, "/")

	unit_id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	invoice_id, err := strconv.Atoi(exploded[5])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	monthly_amount, err := strconv.Atoi(r.Form.Get("monthly_amount"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	
	amount_paid, err := strconv.Atoi(r.Form.Get("amount_paid"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	amount_due, err := strconv.Atoi(r.Form.Get("amount_due"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	d_paid := r.Form.Get("date_paid")
	du_date := r.Form.Get("due_date")

	layout := "2006-01-02"

	date_paid, err := time.Parse(layout, d_paid)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse received date")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	due_date, err := time.Parse(layout, du_date)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse end date")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	Invoice, err := m.DB.GetInvoiceByInvoiceID(invoice_id)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't find invoice")
		http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
		return
	}

	Invoice.InvoiceName = r.Form.Get("invoice_name")
	Invoice.MonthlyAmount = monthly_amount
	Invoice.AmountReceived = amount_paid
	Invoice.DatePaid = date_paid
	Invoice.AmountDue = amount_due
	Invoice.DateDue = due_date
	Invoice.UpdatedAt = time.Now()

	err = m.DB.UpdateInvoice(Invoice)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Updated Invoice successfully")

	http.Redirect(w, r, fmt.Sprintf("/admin/unit-details/%d/show", unit_id), http.StatusSeeOther)
}

//  AdminShowAllUnits displays all units of the owner
func (m *Repository) AdminShowAllUnits(w http.ResponseWriter, r *http.Request) {
	owner_id, ok := m.App.Session.Get(r.Context(), "user_id").(int)

	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get your id from session")
		http.Redirect(w, r, fmt.Sprintf("/admin/all-properties/%d/show", owner_id), http.StatusSeeOther)
		return
	}

	units, err := m.DB.GetAllUnitsByOwnerID(owner_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	intmap := make(map[string]int)
	intmap["owner_id"] = owner_id

	data := make(map[string]interface{})
	data["units"] = units

	render.Template(w, r, "admin-all-owner-units.page.tmpl", &models.TemplateData{
		Data: data,
		IntMap: intmap,
	})
}

// AdminShowAllTenants displays all tenants
func (m *Repository) AdminShowAllTenants(w http.ResponseWriter, r *http.Request) {
	owner_id, ok := m.App.Session.Get(r.Context(), "user_id").(int)

	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get your id from session")
		http.Redirect(w, r, fmt.Sprintf("/admin/all-properties/%d/show", owner_id), http.StatusSeeOther)
		return
	}

	tenants, err := m.DB.GetAllTenantsByOwnerID(owner_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	intmap := make(map[string]int)
	intmap["owner_id"] = owner_id

	data := make(map[string]interface{})
	data["tenants"] = tenants

	render.Template(w, r, "admin-all-tenants.page.tmpl", &models.TemplateData{
		Data: data,
		IntMap: intmap,
	})
}