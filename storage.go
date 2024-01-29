package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateListing(*PetListing) error
	DeleteListingByID(int) error
	UpdateListing(*PetListing) error
	GetListingByID(int) (*PetListing, error)
	GetListings() ([]*PetListing, error)
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
	q := `CREATE TABLE IF NOT EXISTS listing (
			id SERIAL PRIMARY KEY,
			user_id SERIAL NOT NULL,
			name VARCHAR(100) NOT NULL,
			pet_type VARCHAR(50) NOT NULL,
			breed VARCHAR(50),
			sex VARCHAR(10) CHECK (sex IN ('m', 'f')) NOT NULL,
			dob DATE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`

	_, err := cn.db.Exec(q)
	return err
}

func (cn *PostgresDB) CreateListing(listing *PetListing) error {
	query := `INSERT INTO listing (user_id, name, pet_type, breed, sex, dob)
			VALUES ($1, $2, $3, $4, $5, $6);`
	_, err := cn.db.Query(
		query,
		listing.UserId,
		listing.Name,
		listing.PetType,
		listing.Breed,
		listing.Sex,
		listing.Dob)

	if err != nil {
		return err
	}
	return nil
}

func (cn *PostgresDB) DeleteListingByID(id int) error {
	_, err := cn.db.Query("DELETE FROM listing WHERE id = $1;", id)
	return err
}

func (cn *PostgresDB) UpdateListing(*PetListing) error {
	return nil
}

func (cn *PostgresDB) GetListingByID(id int) (*PetListing, error) {
	rows, err := cn.db.Query("SELECT * FROM listing WHERE id = $1", id)
	if err != nil {
		return nil, err

	}
	for rows.Next() {
		return scanListing(rows)
	}
	return nil, fmt.Errorf("listing %d not found", id)
}

func (cn *PostgresDB) GetListings() ([]*PetListing, error) {
	rows, err := cn.db.Query("SELECT * from listing;")
	if err != nil {
		return nil, err
	}

	listings := []*PetListing{}
	for rows.Next() {
		listing, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		listings = append(listings, listing)
	}
	return listings, nil
}

func scanListing(rows *sql.Rows) (*PetListing, error) {
	listing := new(PetListing)
	err := rows.Scan(
		&listing.ID,
		&listing.UserId,
		&listing.Name,
		&listing.PetType,
		&listing.Breed,
		&listing.Sex,
		&listing.Dob,
		&listing.CreatedAt,
		&listing.UpdatedAt)
	return listing, err
}
