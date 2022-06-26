package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/roshanlc/soe-backend/internal/data"
)

// This retrieves student Details
// Handler For GET "/v1/students/:user_id"
func (app *application) showStudentHandler(c *gin.Context) {
	// list of errors
	var errBox data.ErrorBox
	// Check if token matches with provided user ID
	val, token := app.DoesTokenMatchesUserID(c)

	// If user id does not match with token
	if !val {
		errBox.Add(data.AuthorizationErrorResponse("You donot have authorization to access this resource"))
		app.ErrorResponse(c, http.StatusUnauthorized, errBox)
		return
	}

	student, err := app.models.Users.GetStudentDetails(token.UserID)

	// Return error as internal server errror
	if err != nil {
		errBox.Add(data.InternalServerErrorResponse(err.Error()))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}

	// Return the student
	c.JSON(http.StatusOK, gin.H{"student": student})
}

// This update student Details
// Handler For POST "/v1/students/:user_id/update"
func (app *application) updateStudentHandler(c *gin.Context) {
	// list of errors
	var errBox data.ErrorBox
	// Check if token matches with provided user ID
	val, _ := app.DoesTokenMatchesUserID(c)

	// If user id does not match with token
	if !val {
		errBox.Add(data.AuthorizationErrorResponse("You donot have authorization to access this resource"))
		app.ErrorResponse(c, http.StatusUnauthorized, errBox)
		return
	}
	// TODO: add logic
}
