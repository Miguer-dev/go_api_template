package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"errors"
	"net/http"
	"time"

	"go.api.template/internal/validator"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

var (
	ErrTokenRecordNotFoundOrExpiry = errors.New("token record not found or expiry")
	ErrInvalidAuthenticationToken  = func(w http.ResponseWriter) error {
		// inform or remind that we expect to authenticate using a bearer token
		w.Header().Set("WWW-Authenticate", "Bearer")
		return errors.New("invalid or missing authentication token")
	}
	ErrAuthenticationRequired = errors.New("authentication required")
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

type TokenDBConnection struct {
	DB *sql.DB
}

// generate a token instance
func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)

	// Fill the byte slice with random bytes from your operating system's CSPRNG.
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Encode the byte slice to a base-32-encoded string.
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Generate a SHA-256 hash
	token.Hash = HashToken(token.Plaintext)

	return token, nil
}

// hash a plainText
func HashToken(plainText string) []byte {
	hash := sha256.Sum256([]byte(plainText))
	return hash[:]
}

// validate plaintext token has been provided and is exactly 26 bytes long.
func ValidateTokenPlaintext(tokenPlaintext string) *validator.Validator {
	v := &validator.Validator{}

	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")

	return v
}

// generate and insert a token in db
func (t TokenDBConnection) InitToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = t.Insert(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// insert token in db
func (t TokenDBConnection) Insert(token *Token) error {
	query := `
        INSERT INTO tokens (hash, user_id, expiry, scope) 
        VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := t.DB.ExecContext(ctx, query, args...)

	return err
}

// get token by hash and scope, that has not been expiry
func (t TokenDBConnection) GetActiveToken(token *Token) error {
	query := `
		SELECT user_id, expiry FROM tokens
		WHERE hash = $1
		AND scope = $2
		AND expiry > $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := t.DB.QueryRowContext(ctx, query, token.Hash, token.Scope, time.Now()).Scan(
		&token.UserID,
		&token.Expiry)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTokenRecordNotFoundOrExpiry
		} else {
			return err
		}
	}

	return nil
}

// delete all token with specific scope and userID
func (t TokenDBConnection) DeleteAllForUser(scope string, userID int64) error {
	query := `
        DELETE FROM tokens 
        WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := t.DB.ExecContext(ctx, query, scope, userID)

	return err
}
