package data

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"
)

// Wrapper around *sql.DB
type IssuesModel struct {
	DB *sql.DB
}

// struct to hold issue
type Issue struct {
	IssueID   int       `json:"issue_d"`
	Issue     string    `json:"issue"`
	UserID    int       `json:"user_id"`
	UserRole  string    `json:"user_role"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterIssue inserts an issue into db
func (m IssuesModel) RegisterIssue(issue, token string) error {

	query := `INSERT INTO issues (issue, user_id, user_role) VALUES 
	( $1, (SELECT user_id FROM tokens WHERE hash = $2),
	(select roles.name as role FROM users
	 inner join user_roles on users.user_id = user_roles.user_id 
	 inner join roles on roles.role_id = user_roles.role_id 
	 where users.user_id = (SELECT user_id FROM tokens WHERE hash = $2)
	))`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, issue, token)

	if err != nil {
		log.Println(err)
		return err
	}

	// Return nil, that is, issue was inserted successfully
	return nil
}

// MarkAsRead marks an issue as read
// For admins
func (m IssuesModel) MarkAsRead(issueID int) error {

	query := `UPDATE issues SET read = 't' WHERE issue_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.ExecContext(ctx, query, issueID)

	if err != nil {
		log.Println(err)
		return err
	}

	affected, err := rows.RowsAffected()

	if err != nil {
		log.Println(err)
		return err
	}

	// If no issue found by the issue_id provided .i.e 404 Error
	if affected == 0 {
		return ErrRecordNotFound
	}

	// Return nil, that is, issue was inserted successfully
	return nil

}

// Returns a list of all issues
// Supports filter like all, only read, only unread
func (m IssuesModel) GetAllIssues(all, onlyRead bool) (*[]Issue, error) {

	var issues []Issue

	//  issue_id | issue | user_id | user_role | read | created_at
	query := `SELECT issue_id, issue, user_id, user_role, read, created_at 
	FROM issues`

	if !all && onlyRead {
		// if read filter is provided
		query += ` WHERE read = 't'`
	} else if !all && !onlyRead {
		// if unread filter is provided
		query += ` WHERE read = 'f'`
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		switch {
		// No records, but not an error
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecords
		default:
			return nil, err
		}
	}

	defer rows.Close()

	for rows.Next() {

		var temp Issue

		err := rows.Scan(&temp.IssueID,
			&temp.Issue,
			&temp.UserID,
			&temp.UserRole,
			&temp.Read,
			&temp.CreatedAt)

		if err != nil {
			return nil, err
		}

		issues = append(issues, temp)
	}

	// If any  error happened during rows scanning
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &issues, nil
}

// Returns a list of all issues by a user
func (m IssuesModel) GetIssues(userID int) (*[]Issue, error) {

	var issues []Issue

	//  issue_id | issue | user_id | user_role | read | created_at
	query := `SELECT issue_id, issue, user_id, user_role, read, created_at 
	FROM issues WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)

	if err != nil {
		switch {
		// No records, but not an error
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecords
		default:
			return nil, err
		}
	}

	defer rows.Close()

	for rows.Next() {

		var temp Issue

		err := rows.Scan(&temp.IssueID,
			&temp.Issue,
			&temp.UserID,
			&temp.UserRole,
			&temp.Read,
			&temp.CreatedAt)

		if err != nil {
			return nil, err
		}

		issues = append(issues, temp)
	}

	// If any  error happened during rows scanning
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &issues, nil
}
