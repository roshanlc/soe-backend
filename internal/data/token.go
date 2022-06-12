package data

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

// Define constants for the token scope. For now we just define the scope "activation"
// but we'll add additional scopes later as the project goes.
const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

// A struct to hold information about token
type Token struct {
	Plaintext string
	Hash      string
	UserID    int64
	Expiry    time.Time
	Scope     string
}

// A struct to hold info about authentication
type Authentication struct {
	UserID int64     `json:"user_id"` // User ID
	Token  string    `json:"token"`   // Token generated by the system
	Role   string    `json:"role"`    // Role of the user
	Expiry time.Time `json:"expiry"`  // Time at which tokens expires
}

type TokenModel struct {
	DB *sql.DB
}

// Generates a token and returns as a token instance
func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {

	// Create a Token instance containing the user ID, expiry, and scope information.
	// Notice that we add the provided ttl (time-to-live) duration parameter to the
	// current time to get the expiry time?
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// Initialize a zero-valued byte slice with a length of 16 bytes.
	randomBytes := make([]byte, 16)

	// Use the Read() function from the crypto/rand package to fill the byte slice with
	// random bytes from your operating system's CSPRNG. This will return an error if
	// the CSPRNG fails to function correctly.
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Encode the byte slice to a base-32-encoded string and assign it to the token
	// Plaintext field. This will be the token string that we send to the user in their
	// welcome email. They will look similar to this:
	//
	// Y3QMGX3PJ3WLRL2YRTQGQ6KRHU
	//
	// Note that by default base-32 strings may be padded at the end with the =
	// character. We don't need this padding character for the purpose of our tokens, so
	// we use the WithPadding(base32.NoPadding) method in the line below to omit them.
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Generate a MD5 hash of the plaintext token string. Then the hex encoding of this hash
	// will be the value that we store in the `hash` field of our database table. Note that the
	// md5.Sum() function returns an *array* of length 16, so to make it easier to
	// work with we convert it to a slice using the [:] operator before storing it.
	hash := md5.Sum([]byte(token.Plaintext))
	token.Hash = strings.ToUpper(hex.EncodeToString(hash[:])) // Set the hex encoding of hash and upper case them

	return token, nil

}

// Insert into tokens table
func (m TokenModel) Insert(token *Token) error {

	// Construct query to first delete if a token exists already
	// And then insert a new one
	query := `
	INSERT INTO tokens (hash, user_id, expires_at, scope)
	VALUES ($1, $2, $3, $4)`
	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// The New() method is a shortcut which creates a new Token struct and then inserts the
// data in the tokens table.
func (m TokenModel) NewAuthenticationToken(userID int64, ttl time.Duration, scope string) (*Token, error) {

	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)

	return token, err
}

// DeleteAllForUser() deletes all tokens for a specific user and scope.
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
	DELETE FROM tokens
	WHERE scope = $1 AND user_id = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}

// Checks if a token exists in db.
// Returns the token detail if user has that token
func (m TokenModel) LoggedIn(token string) (*Token, error) {

	var authToken Token
	// Construct a query
	query := `
	SELECT hash, user_id, expires_at, scope
	FROM tokens
	WHERE hash = $1 AND scope = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// execute query and scan the result row
	err := m.DB.QueryRowContext(ctx, query, token, ScopeAuthentication).Scan(
		&authToken.Hash,
		&authToken.UserID,
		&authToken.Expiry,
		&authToken.Scope)

	// If any error occurs
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Return the token and nil error
	return &authToken, nil

}

// Logouts a user
// Removes a token from database with authentication scope
func (m TokenModel) LogoutUser(token string) error {

	// Construct a query
	query := `
	DELETE FROM tokens
	WHERE hash = $1 AND scope = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, token, ScopeAuthentication)

	// Incase of errors
	if err != nil {

		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	// No errors, i.e. job was successfull
	return nil
}

// Get a details about a token
func (m TokenModel) GetTokenDetails(token string) (*Token, error) {

	var obj Token
	// Construct a query
	query := `
	SELECT hash,user_id, expires_at, scope FROM tokens
	WHERE hash = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, token).Scan(
		&obj.Hash,
		&obj.UserID,
		&obj.Expiry,
		&obj.Scope,
	)

	// Incase of errors
	if err != nil {

		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// No errors, i.e. job was successfull
	return &obj, nil
}
