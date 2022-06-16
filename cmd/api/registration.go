package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/roshanlc/soe-backend/internal/data"
	"github.com/roshanlc/soe-backend/internal/validator"
)

// registerStudentHandler registers a student
// Handler for POST /v1/students/register
func (app *application) registerStudentHandler(c *gin.Context) {

	// list of errors
	var errBox data.ErrorBox

	var student data.StudentRegistration

	// Bind the data
	err := c.ShouldBindJSON(&student)
	if err != nil {
		errBox.Add(data.BadRequestResponse(err.Error()))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	// Validator for email and password
	v := validator.New()

	data.ValidateEmail(v, student.Email)

	// If errors
	if !v.Valid() {
		errBox.Add(data.CustomErrorResponse("Invalid Email", "Please provide a valid email address."))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	// Validate password if requirements are ok
	data.ValidatePasswordPlaintext(v, student.Password)

	if !v.Valid() {
		errBox.Add(data.CustomErrorResponse("Invalid Password", v.KeyValuePair("password")))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	// Convert plain password into hash
	pw := data.Password{}
	pw.Set(student.Password)

	// Replace the plain password with hash
	student.Password = pw.Hash()

	err = app.models.Users.RegisterStudent(&student)
	if err != nil {

		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			errBox.Add(data.CustomErrorResponse("Duplicate Email", "The provided email is already registered."))
			app.ErrorResponse(c, http.StatusConflict, errBox)
			return
		default:
			errBox.Add(data.InternalServerErrorResponse(err.Error()))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return

		}

	}

	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Account Created", "The account has been created and is currently waiting to be activated by administration"))

	c.JSON(http.StatusCreated, gin.H{"messages": msgBox})

}

// registerTeacherHandler registers a teacher
// Handler for POST /v1/teachers/register
func (app *application) registerTeacherHandler(c *gin.Context) {
	// list of errors
	var errBox data.ErrorBox

	var teacher data.TeacherRegistration

	// Bind the data
	err := c.ShouldBindJSON(&teacher)
	if err != nil {
		errBox.Add(data.BadRequestResponse(err.Error()))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	// Validator for email and password
	v := validator.New()

	data.ValidateEmail(v, teacher.Email)

	// If errors
	if !v.Valid() {
		errBox.Add(data.CustomErrorResponse("Invalid Email", "Please provide a valid email address."))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	// Validate password if requirements are ok
	data.ValidatePasswordPlaintext(v, teacher.Password)

	if !v.Valid() {
		errBox.Add(data.CustomErrorResponse("Invalid Password", v.KeyValuePair("password")))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	// Convert plain password into hash
	pw := data.Password{}
	pw.Set(teacher.Password)

	// Replace the plain password with hash
	teacher.Password = pw.Hash()

	err = app.models.Users.RegisterTeacher(&teacher)

	if err != nil {

		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			errBox.Add(data.CustomErrorResponse("Duplicate Email", "The provided email is already registered."))
			app.ErrorResponse(c, http.StatusConflict, errBox)
			return
		default:
			errBox.Add(data.InternalServerErrorResponse(err.Error()))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return

		}

	}

	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Account Created", "The account has been created and is currently waiting to be activated by administration"))

	c.JSON(http.StatusCreated, gin.H{"messages": msgBox})
}
