package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Struct to hold info about degree levels
type Level struct {
	LevelID int    `json:"level_id"`
	Name    string `json:"name"`
}

// Struct to hold info about program
type Program struct {
	ProgramID  int    `json:"program_id"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Faculty    string `json:"faculty"`
	Level      string `json:"level"`
}

// A wrapper struct around a *sql.DB conn
type ProgramModel struct {
	DB *sql.DB
}

// GetAllLevels Returns the list of all degree levels in db
func (m ProgramModel) GetAllLevels() (*[]Level, error) {

	// Construct the query
	query :=
		`SELECT levels.level_id, levels.name FROM levels`

	// Timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// hold levels
	var levels []Level

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		switch {
		// No level records, which is not exactly an error
		case errors.Is(err, sql.ErrNoRows):

			return nil, ErrNoRecords
		default:
			return nil, err
		}
	}

	// Close rows reading
	defer rows.Close()

	// Loop through rows
	for rows.Next() {
		var temp Level

		err := rows.Scan(&temp.LevelID, &temp.Name)

		if err != nil {
			return nil, err
		}

		// Append the data
		levels = append(levels, temp)
	}

	// Incase any error occured while reading the rows
	if rows.Err() != nil {
		return nil, err
	}

	// Return levels
	return &levels, nil
}

// GetAllProgram Returns the list of all programs in db
func (m ProgramModel) GetAllPrograms(level, department, faculty string) (*[]Program, error) {

	// Construct the query
	query :=
		`SELECT programs.program_id,programs.name, departments.name as dept, faculties.name as faculty,
	levels.name as level FROM programs
	INNER JOIN departments ON departments.department_id = programs.department_id 
	INNER JOIN faculties ON faculties.faculty_id = departments.faculty_id 
	INNER JOIN levels on levels.level_id = programs.level_id
	WHERE (LOWER(levels.name) = LOWER($1) OR $1 = '')
	AND (LOWER(departments.name) = LOWER($2) OR $2 = '')
	AND (LOWER(faculties.name) = LOWER($3) OR $3 = '')`

	// Timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// hold programs
	var programs []Program

	rows, err := m.DB.QueryContext(ctx, query, level, department, faculty)

	if err != nil {
		switch {
		// No program records, which is not exactly an error
		case errors.Is(err, sql.ErrNoRows):

			return nil, ErrNoRecords
		default:
			return nil, err
		}
	}

	// Close rows reading
	defer rows.Close()

	// Loop through rows
	for rows.Next() {
		var temp Program
		err := rows.Scan(&temp.ProgramID,
			&temp.Name,
			&temp.Department,
			&temp.Faculty,
			&temp.Level)

		if err != nil {
			return nil, err
		}

		// Append the data
		programs = append(programs, temp)
	}

	// Incase any error occured while reading the rows
	if rows.Err() != nil {
		return nil, err
	}

	// Return programs
	return &programs, nil
}
