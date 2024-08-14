package models

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.api.template/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail     = errors.New("a user with this email address already exists")
	ErrUserRecordNotFound = errors.New("user record not found")
	ErrInvalidCredentials = errors.New("invalid credential")
	ErrInactiveUser       = errors.New("inactive user")

	AnonymousUser = &User{}
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
}

type UserDBConnection struct {
	DB *sql.DB
}

type password struct {
	plaintext *string
	hash      []byte
}

// calculate the bcrypt hash of a plaintext password, and stores both
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// checks whether the provided plaintext password matches the hashed password
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

// check if a User instance is the AnonymousUser.
func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// validate input user fields
func (u *User) ValidateUser() *validator.Validator {
	v := validator.Validator{}

	v.Check(validator.NotBlank(u.Name), "name", "must be provided")
	v.Check(validator.MaxChars(u.Name, 500), "name", "must not be more than 500 bytes long")

	ValidateEmail(&v, u.Email)
	ValidatePassword(&v, *u.Password.plaintext)

	return &v
}

// validate email
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(validator.NotBlank(email), "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must have email format")
}

// validate password
func ValidatePassword(v *validator.Validator, plainTextPassword string) {
	v.Check(validator.NotBlank(plainTextPassword), "password", "must be provided")
	v.Check(validator.MaxChars(plainTextPassword, 72), "password", "must not be more than 72 bytes long")
	v.Check(validator.MinChars(plainTextPassword, 8), "password", "must not be less than 8 bytes long")
}

// insert an user
func (u UserDBConnection) Insert(user *User) error {
	query := `
        INSERT INTO users (name, email, password_hash, activated) 
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query,
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated).Scan(
		&user.ID,
		&user.CreatedAt,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

// Get user info search by email
func (u UserDBConnection) GetByEmail(email string) (*User, error) {
	query := `
        SELECT id, created_at, name, email, password_hash, activated
        FROM users
        WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrUserRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

// Get user info by Token
func (u UserDBConnection) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
        SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated
        FROM users
        INNER JOIN tokens
        ON users.id = tokens.user_id
        WHERE tokens.hash = $1
        AND tokens.scope = $2 
        AND tokens.expiry > $3`

	args := []any{tokenHash[:], tokenScope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// update a user field
func (u UserDBConnection) UpdateField(id int64, fieldName string, value any) error {
	query := fmt.Sprintf("UPDATE users SET %s = $1 WHERE id = $2", fieldName)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := u.DB.ExecContext(ctx, query, value, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserRecordNotFound
	}

	return nil
}
