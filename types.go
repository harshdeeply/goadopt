package main

import (
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type User struct {
	ID                int       `json:"id"`
	Username          string    `json:"username"`
	EncryptedPassword string    `json:"encrypted_password"`
	CreatedAt         time.Time `json:"created_at"`
}

type Sex string

const (
	Male   Sex = "m"
	Female Sex = "f"
)

type Listing struct {
	ID        int       `json:"id"`
	ListedBy  string    `json:"listed_by"`
	Name      string    `json:"name"`
	PetType   string    `json:"pet_type"`
	Breed     string    `json:"breed"`
	Sex       Sex       `json:"sex"`
	Dob       time.Time `json:"date_of_birth"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateListingRequest struct {
	Name    string    `json:"name"`
	PetType string    `json:"pet_type"`
	Breed   string    `json:"breed"`
	Sex     Sex       `json:"sex"`
	Dob     time.Time `json:"date_of_birth"`
}

type CreateListingResponse struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	PetType   string    `json:"pet_type"`
	Breed     string    `json:"breed"`
	Sex       Sex       `json:"sex"`
	Dob       time.Time `json:"date_of_birth"`
	CreatedAt time.Time `json:"created_at"`
}

func NewListing(listed_by, name, petType, breed string, sex Sex, dob time.Time) *Listing {
	return &Listing{
		ListedBy: listed_by,
		Name:     name,
		PetType:  petType,
		Breed:    breed,
		Sex:      sex,
		Dob:      dob,
	}
}

func NewUser(username, password string) (*User, error) {
	enc_pass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &User{
		Username:          strings.ToLower(username),
		EncryptedPassword: string(enc_pass),
	}, nil
}
