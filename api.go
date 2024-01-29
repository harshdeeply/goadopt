package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func NewAPIServer(listAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/listings", makeHttpHandleFunc(s.handleListings))
	router.HandleFunc("/listings/{id}", withAuth(makeHttpHandleFunc(s.handleListingsByID), s.store))

	log.Println("API is running on port", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

// handles /listings endpoint methods
func (s *APIServer) handleListings(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetListings(w, r)
	} else if r.Method == "POST" {
		return s.handleCreateListing(w, r)
	}
	return fmt.Errorf("method not allowed: %s", r.Method)
}

// handles /listings/{id} endpoint methods
func (s *APIServer) handleListingsByID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetListingByID(w, r)
	} else if r.Method == "DELETE" {
		return s.handleDeleteListing(w, r)
	} else if r.Method == "PUT" {
		return s.handleUpdateListing(w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

// Create new listing
func (s *APIServer) handleCreateListing(w http.ResponseWriter, r *http.Request) error {
	crtLstReq := new(CreateListingRequest)
	if err := json.NewDecoder(r.Body).Decode(&crtLstReq); err != nil {
		return err
	}
	listing := NewListing(crtLstReq.UserId, crtLstReq.Name, crtLstReq.PetType, crtLstReq.Breed, Sex(crtLstReq.Sex), crtLstReq.Dob)
	if err := s.store.CreateListing(listing); err != nil {
		return err
	}

	tokenString, err := createJWTToken(listing)
	if err != nil {
		return err
	}
	fmt.Println("JWT: ", tokenString)
	return WriteJSON(w, http.StatusCreated, listing)
}

// Get all listings
func (s *APIServer) handleGetListings(w http.ResponseWriter, r *http.Request) error {
	listings, err := s.store.GetListings()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, listings)
}

// Update listing by ID
func (s *APIServer) handleUpdateListing(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Get listing by ID
func (s *APIServer) handleGetListingByID(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIdFromReq(r)
	if err != nil {
		return err
	}
	listing, err := s.store.GetListingByID(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, listing)
}

// Delete listing by ID
func (s *APIServer) handleDeleteListing(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIdFromReq(r)
	if err != nil {
		return err
	}
	if err := s.store.DeleteListingByID(id); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

// Parse ID from request utility function
func parseIdFromReq(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id provided %s", idStr)
	}
	return id, nil
}

func withAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("x-auth-token")
		token, err := validateJWTToken(tokenString)
		if err != nil {
			InvalidToken(w)
			return
		}

		if !token.Valid {
			InvalidToken(w)
			return
		}

		listingId, err := parseIdFromReq(r)
		if err != nil {
			InvalidToken(w)
			return
		}
		listing, err := s.GetListingByID(listingId)
		if err != nil {
			InvalidToken(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if listing.ID != int(claims["userId"].(float64)) {
			InvalidToken(w)
			return
		}

		handlerFunc(w, r)
	}
}

func validateJWTToken(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

func createJWTToken(listing *PetListing) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt": 15000,
		"userId":    listing.UserId,
	}
	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{
				Error: err.Error(),
			})
		}
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func InvalidToken(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "invalid token"})
}
