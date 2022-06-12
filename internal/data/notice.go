package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

// A struct to hold information about a notice
type Notice struct {
	ID         int64     `json:"notice_id"`    // Unique identifer for notice
	CreatedAt  time.Time `json:"publish_date"` // Publish date of notice
	Title      string    `json:"title"`        // Title of notice
	Content    string    `json:"content"`      // Content of notice
	MediaLinks []string  `json:"media_links"`  // Attachments included in a notice
	Version    int32     `json:"-"`            // Version, i.e how many modifications have been made
	AddedBy    string    `json:"-"`            // Notice issuer
}

// A NoticeModel struct which wraps a sql.DB connection
type NoticeModel struct {
	DB *sql.DB
}

// Method to retrieve a single notice from database
func (m NoticeModel) Get(id int64) (*Notice, error) {

	// Construct a query for the operation
	query := `SELECT notice_id, created_at, title, content, media_links, version, added_by 
	FROM notices 
	WHERE notice_id = $1 `

	// Create a timeout context of 5 second
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Cancel operation in case of 5 second time out
	defer cancel()

	// empty struct to hold notice
	var notice Notice

	// Use QueryRowContext() to execute the query. This returns a sql.Row
	// containing the result
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&notice.ID,
		&notice.CreatedAt,
		&notice.Title,
		&notice.Content,
		pq.Array(&notice.MediaLinks),
		&notice.Version,
		&notice.AddedBy)

	// Handle any errors. If there was no matching movie found, Scan() will return
	// a sql.ErrNoRows error. We check for this and return our custom ErrRecordNotFound
	// error instead.
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound

		default:
			return nil, err
		}
	}

	// Return the pointer to the notice
	return &notice, nil
}

// Method to retrieve all notices from database
func (m NoticeModel) GetAll(limit int, sort string) ([]*Notice, error) {

	// Construct a query for the operation
	query := fmt.Sprintf(`SELECT notice_id, created_at, title, content, media_links, version, added_by 
	FROM notices 
	ORDER BY created_at %s
	LIMIT $1 `, strings.ToUpper(sort))

	// Create a timeout context of 5 second
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Cancel operation in case of 5 second time out
	defer cancel()

	// Use QueryContext() to execute the query. This returns a sql.Rows resulset
	// containing all the results
	rows, err := m.DB.QueryContext(ctx, query, limit)

	if err != nil {
		return nil, err
	}

	// Importantly, defer a call to rows.Close() to ensure that the resultset is closed
	// before GetAll() returns.
	defer rows.Close()

	// Intialize an empty slice to hold notices
	notices := []*Notice{}

	for rows.Next() {

		// temporary notice struct
		var notice Notice

		// Scan the values from the row into the Notice struct. Again, note that we're
		// using the pq.Array() adapter on the MediaLinks field here.
		err := rows.Scan(&notice.ID,
			&notice.CreatedAt,
			&notice.Title,
			&notice.Content,
			pq.Array(&notice.MediaLinks),
			&notice.Version,
			&notice.AddedBy)

		if err != nil {
			return nil, err
		}

		// Add the notice to the slice
		notices = append(notices, &notice)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notices, nil
}
