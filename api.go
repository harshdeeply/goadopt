package main

import (
	"context"
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
	router.HandleFunc("/signup", makeHttpHandleFunc(s.handleSignup))
	router.HandleFunc("/{username}/listings", makeHttpHandleFunc(s.handleListingsByUsername))
	router.HandleFunc("/listings", makeHttpHandleFunc(s.handleListings)).Methods("GET")
	router.HandleFunc("/listings", withAuth(makeHttpHandleFunc(s.handleListings), s.store)).Methods("POST")
	router.HandleFunc("/listings/{id}", withAuth(makeHttpHandleFunc(s.handleListingsByID), s.store)).Methods("GET", "DELETE")

	log.Println("API is running on port", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

// handles /signup
func (s *APIServer) handleSignup(w http.ResponseWriter, r *http.Request) error {
	// Only POST method is allowed
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed: %s", r.Method)
	}

	crtUserReq := new(CreateUserRequest)
	if err := json.NewDecoder(r.Body).Decode(&crtUserReq); err != nil {
		return err
	}

	user, err := NewUser(crtUserReq.Username, crtUserReq.Password)
	if err != nil {
		return err
	}
	// Check if username taken
	userExists, err := s.store.DoesUserExist(user.Username)
	if err != nil {
		return err
	}

	if userExists {
		return fmt.Errorf("username %s is taken", user.Username)
	}

	// Write user to database
	if err := s.store.CreateUser(user); err != nil {
		return err
	}

	// Sign JWT
	tokenString, err := createJWTToken(user)
	if err != nil {
		return err
	}

	resp := CreateUserResponse{
		Username: user.Username,
		Token:    tokenString,
	}
	return WriteJSON(w, http.StatusCreated, resp)
}

// Get all listings by username
func (s *APIServer) handleListingsByUsername(w http.ResponseWriter, r *http.Request) error {
	username := mux.Vars(r)["username"]
	listings, err := s.store.GetListingsByUsername(username)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, listings)
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
	username := r.Context().Value("username").(string)
	id, err := parseIdFromReq(r)
	if err != nil {
		return err
	}
	ownsListing, err := s.store.CheckOwnership(username, id)
	if err != nil {
		return err
	}
	if ownsListing {
		if r.Method == "GET" {
			return s.handleGetListingByID(w, r)
		} else if r.Method == "DELETE" {
			return s.handleDeleteListing(w, r)
		}
	} else {
		return fmt.Errorf("invalid request")
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

// Create new listing
func (s *APIServer) handleCreateListing(w http.ResponseWriter, r *http.Request) error {
	username := r.Context().Value("username").(string)

	crtLstReq := new(CreateListingRequest)
	if err := json.NewDecoder(r.Body).Decode(&crtLstReq); err != nil {
		return err
	}
	listing := NewListing(username, crtLstReq.Name, crtLstReq.PetType, crtLstReq.Breed, Sex(crtLstReq.Sex), crtLstReq.Dob)
	newListing, err := s.store.CreateListing(listing)
	if err != nil {
		return err
	}
	resp := CreateListingResponse{
		Id:        newListing.ID,
		Name:      newListing.Name,
		Breed:     newListing.Breed,
		PetType:   newListing.PetType,
		Dob:       newListing.Dob,
		Sex:       newListing.Sex,
		CreatedAt: newListing.CreatedAt,
	}
	return WriteJSON(w, http.StatusCreated, resp)
}

// Get all listings
func (s *APIServer) handleGetListings(w http.ResponseWriter, r *http.Request) error {
	listings, err := s.store.GetListings()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, listings)
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

		claims := token.Claims.(jwt.MapClaims)
		ctx := context.WithValue(r.Context(), "username", claims["username"])
		r = r.WithContext(ctx)

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

func createJWTToken(user *User) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt": 15000,
		"username":  user.Username,
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
