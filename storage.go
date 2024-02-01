package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateUser(*User) error
	CreateListing(*Listing) (*Listing, error)
	DeleteListingByID(int) error
	GetListingByID(int) (*Listing, error)
	GetUserByUsername(string) (*User, error)
	GetListings() ([]*Listing, error)
	GetListingsByUsername(string) ([]*Listing, error)
	DoesUserExist(string) (bool, error)
	CheckOwnership(string, int) (bool, error)
}

type PostgresDB struct {
	db *sql.DB
}

// DBConnect initializes a connection to the postgres database
func DBConnect() (*PostgresDB, error) {
	connStr := "user=postgres dbname=postgres password=goadopt sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresDB{
		db: db,
	}, nil
}

// Init creates adoption listing table
func (cn *PostgresDB) Init() error {
	q := `
		CREATE TABLE IF NOT EXISTS listing (
			id SERIAL PRIMARY KEY,
			listed_by VARCHAR(100) NOT NULL,
			name VARCHAR(100) NOT NULL,
			pet_type VARCHAR(50) NOT NULL,
			breed VARCHAR(50),
			sex VARCHAR(10) CHECK (sex IN ('m', 'f')) NOT NULL,
			dob DATE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS "user" (
			id SERIAL PRIMARY KEY,
			username VARCHAR(100) UNIQUE NOT NULL,
			encrypted_password TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT no_spaces_username CHECK (username !~ '\s')
		);
		`

	_, err := cn.db.Exec(q)
	return err
}

func (cn *PostgresDB) CreateUser(user *User) error {
	query := `INSERT INTO "user" (username, encrypted_password)
			VALUES ($1, $2);`
	_, err := cn.db.Query(
		query,
		user.Username,
		user.EncryptedPassword)

	if err != nil {
		return err
	}
	return nil
}

func (cn *PostgresDB) CreateListing(listing *Listing) (*Listing, error) {
	var lastInsertId int64
	query := `INSERT INTO listing (listed_by, name, pet_type, breed, sex, dob)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;`
	err := cn.db.QueryRow(
		query,
		listing.ListedBy,
		listing.Name,
		listing.PetType,
		listing.Breed,
		listing.Sex,
		listing.Dob).Scan(&lastInsertId)
	if err != nil {
		return nil, err
	}
	return cn.GetListingByID(int(lastInsertId))
}

func (cn *PostgresDB) DeleteListingByID(id int) error {
	_, err := cn.db.Query("DELETE FROM listing WHERE id = $1;", id)
	return err
}

func (cn *PostgresDB) GetListingByID(id int) (*Listing, error) {
	rows, err := cn.db.Query("SELECT * FROM listing WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanListing(rows)
	}
	return nil, fmt.Errorf("listing %d not found", id)
}

func (cn *PostgresDB) GetListingsByUsername(username string) ([]*Listing, error) {
	rows, err := cn.db.Query("SELECT * FROM listing WHERE listed_by = $1", username)
	if err != nil {
		return nil, err
	}
	listings := []*Listing{}
	for rows.Next() {
		listing, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		listings = append(listings, listing)
	}
	return listings, nil
}

func (cn *PostgresDB) GetListings() ([]*Listing, error) {
	rows, err := cn.db.Query("SELECT * from listing;")
	if err != nil {
		return nil, err
	}

	listings := []*Listing{}
	for rows.Next() {
		listing, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		listings = append(listings, listing)
	}
	return listings, nil
}

func (cn *PostgresDB) GetUserByUsername(username string) (*User, error) {
	rows, err := cn.db.Query("SELECT * FROM user WHERE username = $1", username)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanUser(rows)
	}
	return nil, fmt.Errorf("username: %s not found", username)
}

func (cn *PostgresDB) DoesUserExist(username string) (bool, error) {
	rows, err := cn.db.Query("SELECT EXISTS(SELECT 1 FROM \"user\" WHERE username = $1)", username)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		var exists bool
		if err := rows.Scan(&exists); err != nil {
			return false, err
		}
		return exists, nil
	}
	return false, fmt.Errorf("unable to check if username exists: %s", username)
}

// Checks if the user with the given username owns the listing with the given listingID.
func (cn *PostgresDB) CheckOwnership(username string, listingID int) (bool, error) {
	var listedBy string
	query := "SELECT listed_by FROM listing WHERE id = $1"
	err := cn.db.QueryRow(query, listingID).Scan(&listedBy)
	if err != nil {
		return false, err
	}
	return username == listedBy, nil
}

func scanListing(rows *sql.Rows) (*Listing, error) {
	listing := new(Listing)
	err := rows.Scan(
		&listing.ID,
		&listing.ListedBy,
		&listing.Name,
		&listing.PetType,
		&listing.Breed,
		&listing.Sex,
		&listing.Dob,
		&listing.CreatedAt,
		&listing.UpdatedAt)
	return listing, err
}

func scanUser(rows *sql.Rows) (*User, error) {
	user := new(User)
	err := rows.Scan(
		&user.ID,
		&user.Username,
		&user.EncryptedPassword,
		&user.CreatedAt)
	return user, err
}
