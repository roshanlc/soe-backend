package data

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/lib/pq"
)

const public_profile = "/v1/profiles/"

var ErrDuplicateProfileID = errors.New("duplicate profile_id")

// Wrapper around *sql.DB
type ProfileModel struct {
	DB *sql.DB
}

// Struct to hold info about teacher's public profile when displaying them all at once
type TeacherProfile struct {
	TeacherID         int      `json:"teacher_id"`
	Name              string   `json:"name"`
	Email             string   `json:"email"`
	ContactNo         string   `json:"contact_no"`
	Profile           string   `json:"profile"`
	Description       string   `json:"description"`
	Academics         []string `json:"academics"`
	Experiences       []string `json:"experiences"`
	Publications      []string `json:"publications"`
	ResearchInterests []string `json:"research_interests"`
}

// Struct to hold info about teacher's public profile
type TeacherProfileDetails struct {
	TeacherID         int         `json:"teacher_id"`
	Name              string      `json:"name"`
	Email             string      `json:"email"`
	ContactNo         string      `json:"contact_no"`
	Profile           string      `json:"profile"`
	Description       string      `json:"description"`
	Academics         []string    `json:"academics"`
	Experiences       []string    `json:"experiences"`
	Publications      []string    `json:"publications"`
	ResearchInterests []string    `json:"research_interests"`
	TeachesAt         []TeachesAt `json:"teaches_at"`
}

/*
type Teacher struct {
	UserID    int64       `json:"user_id"`
	Email     string      `json:"email"`
	TeacherID int64       `json:"teacher_id"`
	Name      string      `json:"name"`
	JoinedAt  time.Time   `json:"joined_at"`
	ContactNo string      `json:"contact_no"`
	Academics []string    `json:"academics"`
	TeachesAt []TeachesAt `json:"teaches_at"`
}

type TeachesAt struct {
	Faculty    string `json:"faculty"`
	Department string `json:"department"`
	Program    string `json:"program"`
	Level      string `json:"level"`
	Semester   int    `json:"semester"`
	Course     string `json:"course"`
	CourseLink string `json:"course_link"`
}
*/

// CreatePublicProfile creates a public profile for a teacher
func (m ProfileModel) CreatePublicProfile(userID int, profileID string, exp, pub, interests []string, descp string) error {

	//   profile_id  |experiences| publications |research_interests|description| teacher_id
	query := `INSERT INTO teacher_profiles (profile_id, experiences, publications, research_interests, description, teacher_id) 
	VALUES ($1, $2, $3, $4, $5, (SELECT teacher_id FROM teachers WHERE user_id = $6))`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, profileID, pq.Array(exp), pq.Array(pub), pq.Array(interests), descp, userID)

	if err != nil {
		switch err.Error() {
		case `pq: duplicate key value violates unique constraint "teacher_profiles_pkey"`:
			return ErrDuplicateProfileID
		case `pq: duplicate key value violates unique constraint "teacher_profiles_teacher_id_key"`:
			return ErrDuplicateEntry

		default:
			log.Println(err)
			return err
		}
	}

	return nil
}

// UpdatePublicProfile updates the public profile of a teacher
func (m ProfileModel) UpdatePublicProfile(userID int, profileID string, exp, pub, interests []string, descp string) error {

	checkQuery := `SELECT COUNT(*) FROM teacher_profiles WHERE teacher_id = (SELECT teacher_id FROM teachers WHERE user_id = $1)`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int
	err := m.DB.QueryRowContext(ctx, checkQuery, userID).Scan(&count)

	if err != nil {
		return err
	}

	// If no record by that id then count = 0
	if count == 0 {
		return ErrRecordNotFound
	}

	//   profile_id  |experiences| publications |research_interests|description| teacher_id
	query := `UPDATE teacher_profiles SET 
	profile_id  = $1,
	experiences = $2,
	publications = $3,
	research_interests = $4,
	description = $5
	WHERE teacher_id = ( SELECT teacher_id FROM teachers WHERE user_id = $6)`

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, profileID, pq.Array(exp), pq.Array(pub), pq.Array(interests), descp, userID)

	// Incase of unknown error
	if err != nil {
		switch err {

		// change not successfull as record does not exist
		// 404 error on the frontend
		case ErrNotUpdated:
			return ErrNotUpdated

		default:
			log.Println(err)
			return err
		}
	}

	affected, err := result.RowsAffected()

	if err != nil {
		log.Println(err)
		return err
	}

	// If affected rows = 0 then something occurred
	if affected == 0 {
		return ErrNotUpdated
	}

	// Return success
	return nil

}

// GetAllPublicProfiles returns the list of public profiles from the db
func (m ProfileModel) GetAllPublicProfiles() (*[]TeacherProfile, error) {

	query :=
		`SELECT users.email, teachers.name, teachers.contact_no, teachers.teacher_id, teachers.academics,
		teacher_profiles.profile_id, teacher_profiles.experiences, teacher_profiles.publications,
		teacher_profiles.research_interests, teacher_profiles.description FROM users
		INNER JOIN teachers ON teachers.user_id = users.user_id
		INNER JOIN teacher_profiles ON teacher_profiles.teacher_id = teachers.teacher_id
		`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)

	// Incase of errors
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecords
		default:
			return nil, err
		}
	}

	defer rows.Close()

	var profiles []TeacherProfile

	// loop through rows
	for rows.Next() {

		// temporary object to hold info
		var temp TeacherProfile
		var profileID string

		err := rows.Scan(&temp.Email,
			&temp.Name,
			&temp.ContactNo,
			&temp.TeacherID,
			pq.Array(&temp.Academics),
			&profileID,
			pq.Array(&temp.Experiences),
			pq.Array(&temp.Publications),
			pq.Array(&temp.ResearchInterests),
			&temp.Description)

		// If some error happened while scanning a row
		if err != nil {
			return nil, err
		}

		temp.Profile = public_profile + profileID

		profiles = append(profiles, temp)

	}

	// If some errors occurred during iteration of rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &profiles, nil
}

// GetAPublicProfile returns a public profile from the db
func (m ProfileModel) GetAPublicProfile(profileID string) (*TeacherProfileDetails, error) {

	basicQuery :=
		`SELECT users.email, teachers.name, teachers.contact_no, teachers.teacher_id, teachers.academics,
		teacher_profiles.profile_id, teacher_profiles.experiences, teacher_profiles.publications,
		teacher_profiles.research_interests, teacher_profiles.description FROM users
		INNER JOIN teachers ON teachers.user_id = users.user_id
		INNER JOIN teacher_profiles ON teacher_profiles.teacher_id = teachers.teacher_id
		WHERE teacher_profiles.profile_id = $1`

	teachesAtQuery := `
	SELECT courses.title, courses.course_code, programs.name,
	departments.name, faculties.name, levels.name , program_courses.semester_id
	FROM teachers
	INNER JOIN teacher_courses ON teachers.teacher_id  = teacher_courses.teacher_id
	INNER JOIN courses ON teacher_courses.course_id = courses.course_id
	INNER JOIN program_courses ON program_courses.course_id  = courses.course_id
	INNER JOIN programs on programs.program_id = program_courses.program_id
	INNER JOIN levels on levels.level_id = programs.level_id
	INNER JOIN departments ON departments.department_id = programs.department_id
	INNER JOIN faculties ON faculties.faculty_id = departments.faculty_id
	INNER JOIN teacher_profiles ON teacher_profiles.teacher_id = teachers.teacher_id
	WHERE teacher_profiles.profile_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var teacher TeacherProfileDetails
	var profile string

	err := m.DB.QueryRowContext(ctx, basicQuery, profileID).Scan(
		&teacher.Email,
		&teacher.Name,
		&teacher.ContactNo,
		&teacher.TeacherID,
		pq.Array(&teacher.Academics),
		&profile,
		pq.Array(&teacher.Experiences),
		pq.Array(&teacher.Publications),
		pq.Array(&teacher.ResearchInterests),
		&teacher.Description)

	// Incase of errors
	if err != nil {
		switch {
		// 404 Not Found
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// set the public profile
	teacher.Profile = public_profile + profile

	ctx1, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Retrieving teaching details
	rows, err := m.DB.QueryContext(ctx1, teachesAtQuery, profileID)

	// If errors
	if err != nil {
		switch {
		// No records here mean that the teacher is not taking any courses for now
		case errors.Is(err, sql.ErrNoRows):
			return &teacher, nil
		default:
			return nil, err
		}

	}
	defer rows.Close()

	for rows.Next() {

		// courses.title, courses.course_code, programs.name,
		// departments.name, faculties.name, levels.name , program_courses.semester_id
		var teachesAt TeachesAt
		var courseCode string
		err := rows.Scan(&teachesAt.Course,
			&courseCode,
			&teachesAt.Program,
			&teachesAt.Department,
			&teachesAt.Faculty,
			&teachesAt.Level,
			&teachesAt.Semester)

		// Incase of any error while reading rows
		if err != nil {
			return nil, err
		}

		// Update the course Link
		teachesAt.CourseLink = courseLink + courseCode

		// Update the teacher instance
		teachingDetails := teacher.TeachesAt
		teachingDetails = append(teachingDetails, teachesAt)
		teacher.TeachesAt = teachingDetails

	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &teacher, nil
}

// Delete method deletes a profile from table
func (m ProfileModel) DeleteProfile(userID int) error {

	// Construct query
	query := `DELETE FROM teacher_profiles WHERE teacher_id = ( SELECT teacher_id  FROM teachers WHERE user_id = $1)`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row, err := m.DB.ExecContext(ctx, query, userID)

	if err != nil {
		return err
	}

	// get the count of rows affected
	affected, err := row.RowsAffected()

	if err != nil {
		return err
	}

	// If affected rows = 0 then the notice did not exist
	if affected == 0 {
		return ErrRecordNotFound
	}

	// Success
	return nil
}
