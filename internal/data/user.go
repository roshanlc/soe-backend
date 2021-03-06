package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/roshanlc/soe-backend/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

// ProfileLinks
var profileLinks = map[string]string{"student": "/v1/students/", "teacher": "/v1/teachers/", "superuser": "/v1/superusers/"}

// course link
const courseLink = "/v1/courses/"

// Create a custom password type which is a struct containing the plaintext and hashed
// versions of the password for a user. The plaintext field is a *pointer* to a string,
// so that we're able to distinguish between a plaintext password not being present in
// the struct at all, versus a plaintext password which is the empty string "".
type Password struct {
	plaintext string
	hash      string
}

// Struct to hold student registration details
type StudentRegistration struct {
	Email      string    `json:"email" binding:"required"`
	Name       string    `json:"name" binding:"required"`
	Password   string    `json:"password" binding:"required,min=8,max=72"`
	SymbolNo   int64     `json:"symbol_no" binding:"required"`
	PURegdNo   string    `json:"pu_regd_no" binding:"required"`
	ContactNo  string    `json:"contact_no" binding:"required"`
	ProgramID  int       `json:"program_id" binding:"required"`
	EnrolledAt time.Time `json:"enrolled_at" binding:"required"`
	Semester   int       `json:"semester" binding:"required"`
}

// Struct to hold teacher registration details
type TeacherRegistration struct {
	Email     string    `json:"email" binding:"required"`
	Name      string    `json:"name" binding:"required"`
	Password  string    `json:"password" binding:"required,min=8,max=72"`
	ContactNo string    `json:"contact_no" binding:"required"`
	Academics []string  `json:"academics" binding:"required"`
	JoinedAt  time.Time `json:"joined_at" binding:"required"`
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
	PURegdNo   string    `json:"pu_regd_no"`
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
	Academics []string    `json:"academics"`
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

// SetHash sets the hash for performing match operation
func (p *Password) SetHash(hash string) {
	p.hash = hash
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
	INNER JOIN teachers ON teachers.user_id = users.user_id
	WHERE users.user_id = $1`,

		// Query for superuser
		"superuser": `SELECT users.user_id, users.email, superusers.name FROM users
	INNER JOIN superusers ON superusers.user_id = users.user_id
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
		&student.PURegdNo,
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

// Get Teacher Details
func (m UserModel) GetTeacherDetails(userID int64) (*Teacher, error) {
	/*

			type Teacher struct {
				UserID            int64       `json:"user_id"`
				Email             string      `json:"email"`
				TeacherID         int64       `json:"teacher_id"`
				Name              string      `json:"name"`
				JoinedAt          time.Time   `json:"joined_at"`
				ContactNo         string      `json:"contact_no"`
				Academics 		  []string    `json:"academics"`
				TeachesAt         []TeachesAt `json:"teaches_at"`
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
	var teacher Teacher

	// Construct query
	basicDetailsQuery := `SELECT users.user_id,users.email, teachers.teacher_id, teachers.name,
	 teachers.joined_at, teachers.contact_no, teachers.academics
	FROM users
	INNER JOIN teachers ON users.user_id = teachers.user_id
	WHERE users.user_id = $1 `

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
	WHERE teachers.teacher_id = $1`

	// 5 sec timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query basic details of a teacher
	err := m.DB.QueryRowContext(ctx, basicDetailsQuery, userID).Scan(
		&teacher.UserID,
		&teacher.Email,
		&teacher.TeacherID,
		&teacher.Name,
		&teacher.JoinedAt,
		&teacher.ContactNo,
		pq.Array(&teacher.Academics))

	// If errors
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}

	}

	// Retrieving teaching details
	rows, err := m.DB.QueryContext(ctx, teachesAtQuery, teacher.TeacherID)

	// If errors
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
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
	// Return the teacher details
	return &teacher, nil
}

// Get SuperUser Details
func (m UserModel) GetSuperUserDetails(userID int64) (*SuperUser, error) {
	/*
		type SuperUser struct {
			UserID      int64  `json:"user_id"`
			Email       string `json:"email"`
			SuperUserID int64  `json:"superuser_id"`
			Name        string `json:"name"`
			AddedBy     string `json:"added_by"`
		}


	*/
	var su SuperUser

	// Construct query
	query := `SELECT users.user_id,users.email, superusers.superuser_id, superusers.name,
	(SELECT superusers.name FROM superusers WHERE superusers.superuser_id = superusers.added_by)
	FROM users
	INNER JOIN superusers ON users.user_id = superusers.user_id
	WHERE users.user_id = $1`

	// 5 sec timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query basic details of a superuser
	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
		&su.UserID,
		&su.Email,
		&su.SuperUserID,
		&su.Name,
		&su.AddedBy,
	)

	// If errors
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}

	}

	// Return the superuser details
	return &su, nil
}

// RegisterStudent registers a student user
func (m UserModel) RegisterStudent(studentReg *StudentRegistration) error {

	query1 := `INSERT INTO users(email, password) VALUES($1, $2) RETURNING user_id`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Role of a student
	var role string = "student"

	// hold user id
	var userID int64

	err := m.DB.QueryRowContext(ctx, query1, studentReg.Email, studentReg.Password).Scan(&userID)

	// Incase of errors
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	// Now, with this user_id we can insert data into students table
	query2 := `INSERT INTO students(name,symbol_no,pu_regd_no,enrolled_at,contact_no,program_id,semester_id,user_id) 
	VALUES( $1, $2, $3, $4, $5, $6, $7, $8)`
	ctx1, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Args value
	args := []interface{}{studentReg.Name, studentReg.SymbolNo, studentReg.PURegdNo, studentReg.EnrolledAt, studentReg.ContactNo,
		studentReg.ProgramID, studentReg.Semester, userID}

	_, err = m.DB.ExecContext(ctx1, query2, args...)

	if err != nil {

		return err
	}

	// Add user_id to user_roles table

	err = RoleModel(m).AddRoleToUser(role, userID)

	if err != nil {
		return err
	}

	// Success
	return nil
}

// RegisterTeacher registers a teacher user
func (m UserModel) RegisterTeacher(teacherReg *TeacherRegistration) error {

	query1 := `INSERT INTO users(email, password) VALUES($1, $2) RETURNING user_id`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Role of a student
	var role string = "teacher"

	// hold user id
	var userID int64

	err := m.DB.QueryRowContext(ctx, query1, teacherReg.Email, teacherReg.Password).Scan(&userID)

	// Incase of errors
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	// Now, with this user_id we can insert data into teachers table
	query2 := `INSERT INTO teachers(name,contact_no,academics,joined_at,user_id) 
	VALUES( $1, $2, $3, $4, $5)`
	ctx1, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Args value
	args := []interface{}{teacherReg.Name, teacherReg.ContactNo, pq.Array(teacherReg.Academics), teacherReg.JoinedAt, userID}

	_, err = m.DB.ExecContext(ctx1, query2, args...)

	if err != nil {
		return err
	}

	// Add user_id to user_roles table

	err = RoleModel(m).AddRoleToUser(role, userID)

	if err != nil {
		return err
	}

	// Success
	return nil
}

// GetPassword retrieves the password of a user
func (m UserModel) GetPassword(userID int64) (string, error) {

	query := `SELECT password FROM users WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var hash string
	err := m.DB.QueryRowContext(ctx, query, userID).Scan(&hash)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return "", ErrRecordNotFound
		default:
			return "", err
		}
	}

	// Return success
	return hash, nil
}

// ChangePassword changes the password of a user
func (m UserModel) ChangePassword(userID int64, newPassHash string) error {

	query := `UPDATE users SET password = $1 WHERE user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, newPassHash, userID)

	if err != nil {
		log.Println(err)
		return err
	}

	affected, err := result.RowsAffected()

	if err != nil {
		log.Println(err)
		return err
	}

	// If affected rows = 0 then old password mis-match
	if affected == 0 {
		return ErrNotUpdated
	}

	// Return success
	return nil
}

// ActivateUser activates the account of a user
func (m UserModel) ActivateUser(token string) error {

	query := `UPDATE users SET activated = 't' WHERE user_id = (SELECT tokens.user_id FROM tokens WHERE hash = $1)`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, token)

	if err != nil {
		log.Println(err)
		return err
	}

	affected, err := result.RowsAffected()

	if err != nil {
		log.Println(err)
		return err
	}

	// If affected rows = 0 then no such token exists
	if affected == 0 {
		return ErrNotUpdated
	}

	// Return success
	return nil
}
