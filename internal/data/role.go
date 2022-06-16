package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Struct to hold info about role
type Role struct {
	RoleID      int64
	Name        string
	Description string
}

// Struct to hold info about user role
type UserRole struct {
	UserID int64
	Role   Role
}

//
type RoleModel struct {
	DB *sql.DB
}

// GetUserRole Returns the role of a user
func (m RoleModel) GetUserRole(userID int64) (*UserRole, error) {

	// Construct query to get role
	query := `
	SELECT roles.role_id, roles.name, roles.description
	FROM user_roles
	INNER JOIN roles ON roles.role_id = user_roles.role_id
	WHERE user_roles.user_id = $1`

	// Create a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var userRole UserRole
	userRole.UserID = userID
	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
		&userRole.Role.RoleID,
		&userRole.Role.Name,
		&userRole.Role.Description,
	)

	// If any error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecords
		default:
			return nil, err
		}
	}

	// Return the user role
	return &userRole, nil
}

// Add user to user_roles table
func (m RoleModel) AddRoleToUser(role string, userID int64) error {

	// Construct query
	query := `INSERT INTO user_roles(user_id,role_id) VALUES ($1, ( SELECT role_id FROM roles WHERE LOWER(roles.name) = LOWER($2) ) )`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, role)

	if err != nil {
		return err
	}

	// Success
	return nil
}
