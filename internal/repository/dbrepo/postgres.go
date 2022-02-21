package dbrepo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Gaussgeek/rentals/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

//InsertNewUser adds a new user in the database
func (m *postgresDBRepo) InsertNewUser(r models.Users) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into users (first_name, last_name, email, password, created_at, updated_at)
			values ($1, $2, $3, $4, $5, $6)`

	_, err := m.DB.ExecContext(ctx, stmt,
		r.FirstName,
		r.LastName,
		r.Email,
		r.Password,
		r.CreatedAt,
		r.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

// GetUserByID returns a user by id
func (m *postgresDBRepo) GetUserByID(id int) (models.Users, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, first_name, last_name, email, is_email_verified, password, access_level, created_at, updated_at
			from users where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.Users
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.IsEmailVerified,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}

// UpdateUser updates a user in the database for a given ID
func (m *postgresDBRepo) UpdateUser(u models.Users) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update users set first_name = $1, last_name = $2, email = $3, phone =$4 access_level = $5, updated_at = $6
		where id = $7
`

	_, err := m.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.Phone,
		u.AccessLevel,
		time.Now(),

		u.ID,
	)

	if err != nil {
		return err
	}

	return nil
}


// Authenticate authenticates a user
func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "select id, password from users where email = $1", email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}

func (m *postgresDBRepo) AddNewTokenToUser(id int, s string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	 
	query := `
		update users  set token = $1, token_expiry = $2 where id = $3
`
	t := time.Now().Add(2*time.Hour)

	_, err := m.DB.ExecContext(ctx, query,
		s,
		t,
		id,
	)

	if err != nil {
		return err
	}

	return nil
}

// DeleteUserByID deletes a user from the database
func (m *postgresDBRepo) DeleteUserByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	query := `delete from users where id = $1`

	_, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}
	return nil
}

// AddNewProperty adds a new property to the database
func (m *postgresDBRepo) AddNewProperty(p models.Property) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into property (property_name, property_location, owner_id, created_at, updated_at)
			 values ($1, $2, $3, $4, $5)`
	
	 _, err := m.DB.ExecContext(ctx, query,
				p.PropertyName,
				p.PropertyLocation,
				p.OwnerID,
				p.CreatedAt,
				p.UpdatedAt,
	 )
	if err != nil {
		return err
	}

	return nil
}

// GetPropertiesByOwnwerID returns all properties with the same owner_id
func (m *postgresDBRepo)GetPropertiesByOwnwerID(id int) ([]models.Property, error)  {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var properties []models.Property

	query := `select id, property_name, property_location, owner_id, created_at, updated_at
			from property where owner_id = $1`
	
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Property
		err := rows.Scan(
			&p.ID,
			&p.PropertyName,
			&p.PropertyLocation,
			&p.OwnerID,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		properties = append(properties, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return properties, nil
}

func (m *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	if email == "me@here.ca" {
		return 1, "", nil
	}
	return 0, "", errors.New("some error")
}

func (m *postgresDBRepo) GetPropertyByPropertyID(id int) (models.Property, error)  {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, property_name, property_location, owner_id, created_at, updated_at
			from property where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var p models.Property
	err := row.Scan(
				&p.ID,
				&p.PropertyName,
				&p.PropertyLocation,
				&p.OwnerID,
				&p.CreatedAt,
				&p.UpdatedAt,
			)
		
	if err != nil {
			return p, err
 		  }
		
	return p, nil
}

func (m *postgresDBRepo) InsertNewUnit(u models.Unit) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into unit (unit_name, property_id,  occupancy_status, created_at, updated_at) 
			values ($1, $2, $3, $4, $5)`

	_, err := m.DB.ExecContext(ctx, stmt, 
			u.UnitName,
			u.PropertyID,
			u.OccupancyStatus,
			u.CreatedAt,
			u.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) UpdateUnitDetails(u models.Unit) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update unit set unit_name =$1, occupancy_status=$2, updated_at=$3
			where id = $4`
	
	_, err := m.DB.ExecContext(ctx, stmt, 
				u.UnitName,
				u.OccupancyStatus,
				u.UpdatedAt,

				u.ID,
	)
	if err != nil {
		return err		 
	}

	return nil
}

func (m *postgresDBRepo) InsertNewTenant(u models.Tenant) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into tenant 
			(first_name, last_name, email, phone, other_phone, alternate_contact_name, alternate_contact_phone, risk_id, unit_id,
				date_of_occupancy, exit_date, invoice_id, created_at, updated_at)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	
	_, err := m.DB.ExecContext(ctx, stmt, 
				u.FirstName,
				u.LastName,
				u.Email,
				u.Phone,
				u.OtherPhone,
				u.AlternateContactPersonName,
				u.AlternateContactPersonPhone,
				u.RiskID,
				u.UnitID,
				u.DateOfOccupancy,
				u.ExitDate,
				u.InvoiceID,
				u.CreatedAt,
				u.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) GetUnitsByPropertyID(id int) ([]models.Unit, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var units []models.Unit

	query := `select id, unit_name, property_id, occupancy_status,
			  created_at, updated_at from unit where property_id = $1`
	
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
			  }
	defer rows.Close()
		  
	for rows.Next() {
		var u models.Unit
		err := rows.Scan(
				  &u.ID,
				  &u.UnitName,
				  &u.PropertyID,
				  &u.OccupancyStatus,
				  &u.CreatedAt,
				  &u.UpdatedAt,
				  )
    	if err != nil {
			return nil, err
		}
		  
		units = append(units, u)
			
	}

	if err = rows.Err(); err != nil {
	 return nil, err
 	}

	return units, nil
}

func (m *postgresDBRepo) GetAllUnitsByOwnerID(id int) ([]models.Unit, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var units []models.Unit
	
	query := `select all u.id, u.unit_name , u.property_id , u.occupancy_status, u.created_at , u.updated_at 
			from unit u
			where property_id in 
			(select all id from property p where owner_id = $1)`

	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	 }
	defer rows.Close()
				  
	for rows.Next() {
				var u models.Unit
				err := rows.Scan(
						  &u.ID,
						  &u.UnitName,
						  &u.PropertyID,
						  &u.OccupancyStatus,
						  &u.CreatedAt,
						  &u.UpdatedAt,
						  )
				if err != nil {
					return nil, err
					}
				  
		units = append(units, u)
					
	}
		
	if err = rows.Err(); err != nil {
			 return nil, err
	}
		
	return units, nil
}

func (m *postgresDBRepo) GetUnitByUnitID(id int) (models.Unit, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, unit_name, property_id, occupancy_status, created_at, updated_at
			from unit where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.Unit

	err := row.Scan(
		&u.ID,
		&u.UnitName,
		&u.PropertyID,
		&u.OccupancyStatus,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// there were no rows, but otherwise no error occurred
			return u, nil
		} else {
		return u, err
		}
	}

	return u, nil
}

func (m *postgresDBRepo) GetTenantByUnitID(id int) (models.Tenant, error){
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, first_name, last_name, email, phone, other_phone, alternate_contact_name, alternate_contact_phone, 
				risk_id, unit_id, date_of_occupancy, exit_date, invoice_id,
				created_at, updated_at
			from tenant where unit_id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.Tenant

	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Phone,
		&u.OtherPhone,
		&u.AlternateContactPersonName,
		&u.AlternateContactPersonPhone,
		&u.RiskID,
		&u.UnitID,
		&u.DateOfOccupancy,
		&u.ExitDate,
		&u.InvoiceID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// there were no rows, but otherwise no error occurred
			return u, nil
		} else {
		return u, err
		}
	}

	return u, nil
}

func (m *postgresDBRepo) GetInvoiceByUnitID(id int) (models.Invoice, error){
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, invoice_name, unit_id, monthly_amount, amount_paid, date_paid,
			amount_due, due_date, created_at, updated_at
			from invoice where unit_id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.Invoice

	err := row.Scan(
		&u.ID,
		&u.InvoiceName,
		&u.UnitID,
		&u.MonthlyAmount,
		&u.AmountReceived,
		&u.DatePaid,
		&u.AmountDue,
		&u.DateDue,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// there were no rows, but otherwise no error occurred
			return u, nil
		} else {
		return u, err
		}
	}

	return u, nil
}

func (m *postgresDBRepo) GetExpenseByUnitID(id int) (models.Expenses, error){
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, expense_name, unit_id, amount_paid, date_paid, narration, amount_due,
			due_date, created_at, updated_at
			from expenses where unit_id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.Expenses

	err := row.Scan(
		&u.ID,
		&u.ExpenseName,
		&u.UnitID,
		&u.AmountPaid,
		&u.DatePaid,
		&u.Narration,
		&u.AmountDue,
		&u.DueDate,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// there were no rows, but otherwise no error occurred
			return u, nil
		} else {
		return u, err
		}
	}

	return u, nil
}

func (m *postgresDBRepo) UpdateTenant(u models.Tenant) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update tenant set 
				first_name = $1, last_name=$2, email=$3, phone=$4, other_phone=$5, alternate_contact_name=$6, alternate_contact_phone=$7, risk_id=$8, unit_id=$9,
				date_of_occupancy=$10, exit_date=$11, invoice_id=$12, created_at=$13, updated_at=$14
				where id = $15`

	_, err := m.DB.ExecContext(ctx, query, 
		u.FirstName,
		u.LastName,
		u.Email,
		u.Phone,
		u.OtherPhone,
		u.AlternateContactPersonName,
		u.AlternateContactPersonPhone,
		u.RiskID,
		u.UnitID,
		u.DateOfOccupancy,
		u.ExitDate,
		u.InvoiceID,
		u.CreatedAt,
		u.UpdatedAt,
		
		u.ID,
	
		)	

		if err != nil {
			return err
		}
	
		return nil
}


func (m *postgresDBRepo) DeleteTenant(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "delete from tenant where id = $1"

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) InsertNewExpense(u models.Expenses) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into expenses (expense_name, unit_id, amount_paid, date_paid, 
			narration, amount_due, due_date, created_at, updated_at)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := m.DB.ExecContext(ctx, stmt,
				u.ExpenseName,
				u.UnitID,
				u.AmountPaid,
				u.DatePaid,
				u.Narration,
				u.AmountDue,
				u.DueDate,
				u.CreatedAt,
				u.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) UpdateExpense(u models.Expenses) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update expenses set expense_name = $1, amount_paid= $2, date_paid= $3,
				amount_due= $4, due_date= $5, updated_at= $6
				where id = $7`

	_, err := m.DB.ExecContext(ctx, query, 
					u.ExpenseName,
					u.AmountPaid,
					u.DatePaid,
					u.AmountDue,
					u.DueDate,
					u.UpdatedAt, 
					
					u.ID,
				
					)	
			
	if err != nil {
		return err
	}
				
	return nil
}

func (m *postgresDBRepo) InsertNewInvoice(u models.Invoice) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into invoice (invoice_name, unit_id, monthly_amount, amount_paid, date_paid, 
			amount_due, due_date, created_at, updated_at)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := m.DB.ExecContext(ctx, stmt,
				u.InvoiceName,
				u.UnitID,
				u.MonthlyAmount,
				u.AmountReceived,
				u.DatePaid,
				u.AmountDue,
				u.DateDue,
				u.CreatedAt,
				u.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) GetInvoicesByUnitID(id int) ([]models.Invoice, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var invoices []models.Invoice

	query := `select id, invoice_name, unit_id, monthly_amount, amount_paid, date_paid, amount_due, due_date, created_at, updated_at
			from invoice where unit_id = $1`
	
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Invoice
		err := rows.Scan(
			&p.ID,
			&p.InvoiceName,
			&p.UnitID,
			&p.MonthlyAmount,
			&p.AmountReceived,
			&p.DatePaid,
			&p.AmountDue,
			&p.DateDue,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		invoices = append(invoices, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return invoices, nil
}

func (m *postgresDBRepo) GetInvoiceByInvoiceID(id int) (models.Invoice, error){
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, invoice_name, unit_id, monthly_amount, amount_paid, date_paid, amount_due, due_date, created_at, updated_at
			from invoice where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.Invoice

	err := row.Scan(
		&u.ID,
		&u.InvoiceName,
		&u.UnitID,
		&u.MonthlyAmount,
		&u.AmountReceived,
		&u.DatePaid,
		&u.AmountDue,
		&u.DateDue,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// there were no rows, but otherwise no error occurred
			return u, nil
		} else {
		return u, err
		}
	}

	return u, nil
}

func (m *postgresDBRepo) UpdateInvoice(u models.Invoice) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update invoice set invoice_name =$1, monthly_amount= $2, amount_paid=$3, date_paid= $4, amount_due=$5, due_date=$6, updated_at=$7
			 where id = $8`

	_, err := m.DB.ExecContext(ctx, query, 
				u.InvoiceName,
				u.MonthlyAmount,
				u.AmountReceived,
				u.DatePaid,
				u.AmountDue,
				u.DateDue,
				u.UpdatedAt,
				
				u.ID,
	)

	if err != nil {
		return err
	}
				
	return nil
}