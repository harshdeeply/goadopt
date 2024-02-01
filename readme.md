# Pet Adoption App API

This project is an API written in Golang for a pet adoption app. It utilizes PostgreSQL as its database, JWT for user authentication, goroutines for concurrency, and runs within a Docker container managed by Docker Compose.

## Endpoints

### GET /listings

- Returns all listings.
- No authentication token needed in the header.

### POST /signup

- Registers a new user and returns a JWT token.
- Request body should be in the following format:
  ```json
  {
    "username": "johnwick",
    "password": "superComplexPassword"
  }
  ```

### POST /listings

- Creates a new listing.
- Requires a valid JWT token in the header as `x-auth-token: authentication_token`.
- Request body should be in the following format:
  ```json
  {
    "name": "Jordan",
    "pet_type": "Dog",
    "breed": "German Shephard",
    "sex": "m",
    "date_of_birth": "2020-01-10T00:00:00Z"
  }
  ```

### GET /listing/{id}

- Retrieves the listing with the specified ID.
- Requires a valid JWT token in the header.
- Only allows access to listings owned by the authenticated user, otherwise returns an error.

## Authentication

- JWT (JSON Web Tokens) are used for user authentication.
- Users can obtain a token by signing up using the `/signup` endpoint.
- The obtained token should be included in the header of subsequent authenticated requests as `x-auth-token: token`.

## Dependencies

- PostgreSQL is used as the database.
- Golang's standard library is utilized for the majority of functionalities.
- JWT library is used for user authentication.

## Running the Application

To run the application using Docker Compose:

1. Make sure you have Docker and Docker Compose installed on your system.
2. Clone this repository.
3. Navigate to the project directory.
4. Run the following command to build and start the containers:
   ```
   docker-compose up --build
   ```

The API should now be accessible at `http://localhost:8080`.
