package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/roshanlc/soe-backend/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

// ProfileLinks
var profileLinks = map[string]string{"student": "/v1/students/", "teacher": "/v1/teachers/", "superuser": "/v1/superusers/"}

// Create a custom password type which is a struct containing the plaintext and hashed
// versions of the password for a user. The plaintext field is a *pointer* to a string,
// so that we're able to distinguish between a plaintext password not being present in
// the struct at all, versus a plaintext password which is the empty string "".
type Password struct {
	plaintext string
	hash      string
}

// struct to hold details of user
type User struct {
	UserID    int64    `json:"user_id"`
	Email     string   `json:"email"`
	Password  Password `json:"-"` // Dash "-" means this field will not be exported to client as json
	Activated bool     `json:"-"`
	Expired   bool     `json:"-"`
	Version   int      `json:"-"`
}

// Struct to hold user input for login endpoint
type LoginInput struct {
	Email    string `json:"email" binding:"required"`                 // email of user
	Password string `json:"password" binding:"required,min=8,max=72"` // password
}

// struct to hold basic details of User
type UserDetails struct {
	UserID  int64  `json:"user_id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Role    string `json:"role"`
	Profile string `json:"profile"` // Profile , like "/v1/students/:user_id","/v1/teacher/:user_id"
}

// Struct to hold details of student
type Student struct {
	UserID     int64     `json:"user_id"`
	Email      string    `json:"email"`
	StudentID  int64     `json:"student_id"`
	Name       string    `json:"name"`
	SymbolNo   int64     `json:"symbol_no"`
	PURedgNo   string    `json:"pu_regd_no"`
	EnrolledAt time.Time `json:"enrolled_at"`
	Faculty    string    `json:"faculty"`
	Department string    `json:"department"`
	Program    string    `json:"program"`
	Level      string    `json:"level"`
	Semester   int       `json:"semester"`
}

// A substruct for teacher group
type TeachesAt struct {
	Faculty    string `json:"faculty"`
	Department string `json:"department"`
	Program    string `json:"program"`
	Level      string `json:"level"`
	Semester   int    `json:"semester"`
	Course     string `json:"course"`
	CourseLink string `json:"course_link"`
}

// Struct to hold details of teacher
type Teacher struct {
	UserID    int64       `json:"user_id"`
	Email     string      `json:"email"`
	TeacherID int64       `json:"teacher_id"`
	Name      string      `json:"name"`
	JoinedAt  time.Time   `json:"joined_at"`
	ContactNo string      `json:"contact_no"`
	TeachesAt []TeachesAt `json:"teaches_at"`
}

//  A struct to hold info about superuser
type SuperUser struct {
	UserID      int64  `json:"user_id"`
	Email       string `json:"email"`
	SuperUserID int64  `json:"superuser_id"`
	Name        string `json:"name"`
	AddedBy     string `json:"added_by"`
}

// Wrapper around *sql.DB connection
type UserModel struct {
	DB *sql.DB
}

// Return Hash of password
func (p *Password) Hash() string {
	return p.hash
}

// The Set() method calculates the bcrypt hash of a plaintext password, and stores both
// the hash and the plaintext versions in the struct.
func (p *Password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = plaintextPassword
	p.hash = string(hash)
	return nil
}

// The Matches() method checks whether the provided plaintext password matches the
// hashed password stored in the struct, returning true if it matches and false
// otherwise.
func (p *Password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plaintextPassword))
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

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

// Returns the user detail by email
func (m UserModel) GetByEmail(email string) (*User, error) {

	// Construct a query
	query := `SELECT user_id, email, password, activated, expired, version
	FROM users
	WHERE LOWER(email) = LOWER($1)`

	var user User

	// Create a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Cancel the transaction if timeout
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.UserID,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Expired,
		&user.Version)

	// If any errors
	if err != nil {
		switch {
		// If no record found, i.e. invalid email
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil

}

// Get User Details
func (m UserModel) GetUserDetails(userID int64) (*UserDetails, error) {

	/*
		UserID  int64  `json:"user_id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Role    string `json:"role"`
		Profile string `json:"profile"`
	*/

	var userTypeQuery = map[string]string{
		// Query for student
		"student": `SELECT users.user_id, users.email, students.name FROM users
	INNER JOIN students ON students.user_id = users.user_id
	WHERE users.user_id = $1`,
		// Query for teacher
		"teacher": `SELECT users.user_id, users.email, teachers.name FROM users
	INNER JOIN students ON students.user_id = users.user_id
	WHERE users.user_id = $1`,

		// Query for superuser
		"superuser": `SELECT users.user_id, users.email, superusers.name FROM users
	INNER JOIN students ON students.user_id = users.user_id
	WHERE users.user_id = $1`,
	}

	var user UserDetails

	roleModel := RoleModel(m)

	role, err := roleModel.GetUserRole(userID)

	// If any errors
	if err != nil {
		return nil, err
	}

	// Set the role name
	user.Role = role.Role.Name

	// Corresponding query
	query := userTypeQuery[role.Role.Name]

	// Create time out context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.Email,
		&user.Name,
	)

	// If any error
	if err != nil {
		return nil, err
	}

	// Set the user profile link for specific details
	user.Profile = fmt.Sprintf("%s%d", profileLinks[user.Role], (user.UserID))

	return &user, nil
}

// Get Student Details
func (m UserModel) GetStudentDetails(userID int64) (*Student, error) {
	/*

		type Student struct {
			UserID     int64     `json:"user_id"`
			Email      string    `json:"email"`
			StudentID  int64     `json:"student_id"`
			Name       string    `json:"name"`
			SymbolNo   int64     `json:"symbol_no"`
			PURedgNo   string    `json:"pu_regd_no"`
			EnrolledAt time.Time `json:"enrolled_at"`
			Faculty    string    `json:"faculty"`
			Department string    `json:"department"`
			Program    string    `json:"program"`
			Level      string    `json:"level"`
			Semester   int       `json:"semester"`
		}
	*/

	var student Student

	// Construct query
	query := `SELECT users.user_id, users.email, students.student_id, students.name,
	students.symbol_no, students.pu_regd_no, students.enrolled_at,
	faculties.name, departments.name, programs.name, levels.name, students.semester_id
	FROM users 
	INNER JOIN students ON students.user_id = users.user_id
	INNER JOIN programs ON programs.program_id = students.program_id
	INNER JOIN departments ON departments.department_id = programs.department_id
	INNER JOIN faculties ON faculties.faculty_id = departments.faculty_id
	INNER JOIN levels ON levels.level_id = programs.level_id
	WHERE users.user_id = $1 `

	// 5 sec timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
		&student.UserID,
		&student.Email,
		&student.StudentID,
		&student.Name,
		&student.SymbolNo,
		&student.PURedgNo,
		&student.EnrolledAt,
		&student.Faculty,
		&student.Department,
		&student.Program,
		&student.Level,
		&student.Semester)

	// If errors
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}

	}

	// Return the student details
	return &student, nil
}

/*

// Get Teacher Details
func (m UserModel) GetTeacherDetails(userID int64) (*Student, error) {
	/*

		type Teacher struct {
			UserID            int64       `json:"user_id"`
			Email             string      `json:"email"`
			TeacherID         int64       `json:"teacher_id"`
			Name              string      `json:"name"`
			JoinedAt          time.Time   `json:"joined_at"`
			ContactNo         string      `json:"contact_no"`
			TeachesAt         []TeachesAt `json:"teaches_at"`
		}


	var teacher Teacher

	// Construct query
	query := `SELECT users.user_id, users.email, teacher.student_id, students.name,
	students.symbol_no, students.pu_regd_no, students.enrolled_at,
	faculties.name, departments.name, programs.name, levels.name, students.semester_id
	FROM users
	INNER JOIN students ON students.user_id = users.user_id
	INNER JOIN programs ON programs.program_id = students.program_id
	INNER JOIN departments ON departments.department_id = programs.department_id
	INNER JOIN faculties ON faculties.faculty_id = departments.faculty_id
	INNER JOIN levels ON levels.level_id = programs.level_id
	WHERE users.user_id = $1 `

	// 5 sec timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
		&student.UserID,
		&student.Email,
		&student.StudentID,
		&student.Name,
		&student.SymbolNo,
		&student.PURedgNo,
		&student.EnrolledAt,
		&student.Faculty,
		&student.Department,
		&student.Program,
		&student.Level,
		&student.Semester)

	// If errors
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}

	}

	// Return the student details
	return &student, nil
}

*/
