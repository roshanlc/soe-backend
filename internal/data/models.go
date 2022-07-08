package data

import (
	"database/sql"
	"errors"
)

// Define a custom ErrRecordNotFound error. We'll return this from our Get() method when
// looking up a object that doesn't exist in our database.
var (
	ErrRecordNotFound      = errors.New("record not found")               // specific record not found
	ErrEditConflict        = errors.New("edit conflict")                  // conflict occured while editing
	ErrNoRecords           = errors.New("no records")                     // view is returned empty
	ErrDuplicateEmail      = errors.New("duplicate email")                // Incase of duplicate email
	ErrOldPasswordMisMatch = errors.New("old password does not match")    // Incase of old password mismatch during password change
	ErrNotUpdated          = errors.New("the change was not successfull") // Incase of failure while changing info
	ErrDuplicateEntry      = errors.New("duplicate entry denied")         // Incase of duplicate entry
)

// All models within a single wrapper struct
type Models struct {
	Notices  NoticeModel  // Notice model
	Courses  CourseModel  // Course Model
	Users    UserModel    // User model
	Tokens   TokenModel   // Token Model
	Roles    RoleModel    // Role Model
	Programs ProgramModel // Programs model
	Schedule ScheduleModel
	Issues   IssuesModel
}

// Returns a models object
func NewModels(db *sql.DB) Models {
	return Models{
		Notices:  NoticeModel{DB: db}, // NoticeModel
		Courses:  CourseModel{DB: db}, // Course Model
		Users:    UserModel{DB: db},   // User Model
		Tokens:   TokenModel{DB: db},  // Token Model
		Roles:    RoleModel{DB: db},
		Programs: ProgramModel{DB: db},
		Schedule: ScheduleModel{DB: db},
		Issues:   IssuesModel{DB: db},
	}
}
