package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/roshanlc/soe-backend/internal/data"
	"github.com/roshanlc/soe-backend/internal/validator"
)

const activatedHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta http-equiv="X-UA-Compatible" content="IE=edge" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Account Activated</title>
    <style>
      h3 {
        text-align: center;
        font-family: "Gill Sans", "Gill Sans MT", Calibri, "Trebuchet MS",
          sans-serif;
      }
    </style>
  </head>
  <body>
    <h3>Your account has been activated!</h3>
  </body>
</html>
`

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
	token, err := app.models.Tokens.GenAndInsertActivationToken(student.Email, 24*time.Hour)
	if err != nil {
		log.Println(err)
		errBox.Add(data.InternalServerErrorResponse(err.Error()))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return

	}
	link := app.config.Domain + "/v1/users/activate?token=" + token.Hash

	mailDetails := MailingContent{from: app.config.Mail.Sender, to: student.Email,
		subject: "Activation Link", content: "Please click on the following <a href = \"" + link + "\">activation link</a> to activate your account."}

	go app.mailHandler.SendMail(&mailDetails)

	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Account Created", "Please check your email for activation link. The activation link is valid for 24 hours."))

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

	token, err := app.models.Tokens.GenAndInsertActivationToken(teacher.Email, 24*time.Hour)
	if err != nil {
		log.Println(err)
		errBox.Add(data.InternalServerErrorResponse(err.Error()))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return

	}
	link := app.config.Domain + "/v1/users/activate?token=" + token.Hash

	mailDetails := MailingContent{from: app.config.Mail.Sender, to: teacher.Email,
		subject: "Activation Link",
		content: "Please click on the following <a href = \"" + link + "\">activation link</a> to activate your account."}

	go app.mailHandler.SendMail(&mailDetails)

	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Account Created", "Please check your email for activation link. The activation link is valid for 24 hours."))

	c.JSON(http.StatusCreated, gin.H{"messages": msgBox})
}

// activateUserHandler activates a user account
// Handler for GET /v1/users/activate
func (app *application) activateUserHandler(c *gin.Context) {
	// list of errors
	var errBox data.ErrorBox

	tokenVal, exists := c.GetQuery("token")

	if !exists {
		errBox.Add(data.BadRequestResponse("A query paramter (\"token\") is missing."))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	err := app.models.Users.ActivateUser(tokenVal)
	if err != nil {
		switch err {
		// no such token exists
		case data.ErrNotUpdated:
			errBox.Add(data.ResourceNotFoundResponse("Expired or non-existing token value."))
			app.ErrorResponse(c, http.StatusNotFound, errBox)
			return

		default:
			errBox.Add(data.InternalServerErrorResponse("The server had problems while processing this request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}
	}

	err = app.models.Tokens.DeleteByToken(tokenVal, data.ScopeActivation)

	if err != nil {
		fmt.Println(err)
		errBox.Add(data.InternalServerErrorResponse("The server had problems while processing this request."))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(activatedHTML))

}
