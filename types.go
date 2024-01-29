package main

import (
	"time"
)

type Sex string

const (
	Male   Sex = "m"
	Female Sex = "f"
)

type PetListing struct {
	ID        int       `json:"id"`
	UserId    int       `json:"user_id"`
	Name      string    `json:"name"`
	PetType   string    `json:"pet_type"`
	Breed     string    `json:"breed"`
	Sex       Sex       `json:"sex"`
	Dob       time.Time `json:"date_of_birth"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateListingRequest struct {
	UserId  int       `json:"user_id"`
	Name    string    `json:"name"`
	PetType string    `json:"pet_type"`
	Breed   string    `json:"breed"`
	Sex     string    `json:"sex"`
	Dob     time.Time `json:"date_of_birth"`
}

func NewListing(user_id int, name, petType, breed string, sex Sex, dob time.Time) *PetListing {
	return &PetListing{
		// ID:          rand.Intn(10000),
		UserId:  user_id,
		Name:    name,
		PetType: petType,
		Breed:   breed,
		Sex:     sex,
		Dob:     dob,
	}
}
