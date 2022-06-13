// This contains handlers for endpoints related to user,
// such as user details, user roles, and others.
package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/roshanlc/soe-backend/internal/data"
)

// This retrieves user's basic details
// Handler for GET "/v1/users/:user_id"
func (app *application) showUserHandler(c *gin.Context) {

	// list of errors
	var errBox data.ErrorBox

	// Check if token matches with provided user ID
	val, token := app.DoesTokenMatchesUserID(c)

	// If user id does not match with token
	if !val {
		return
	}

	user, err := app.models.Users.GetUserDetails(token.UserID)

	// Return error as internal server errror
	if err != nil {
		errBox.Add(data.InternalServerErrorResponse(err.Error()))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}

	// Return the student
	c.JSON(http.StatusOK, gin.H{"user": user})

}

// This retrieves student Details
// Handler For GET "/v1/students/:user_id"
func (app *application) showStudentHandler(c *gin.Context) {
	// list of errors
	var errBox data.ErrorBox
	// Check if token matches with provided user ID
	val, token := app.DoesTokenMatchesUserID(c)

	// If user id does not match with token
	if !val {
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

// Show teacher handler
// Handler For GET "/v1/teachers/:user_id"
func (app *application) showTeacherHandler(c *gin.Context) {
	// list of errors
	var errBox data.ErrorBox
	// Check if token matches with provided user ID
	val, token := app.DoesTokenMatchesUserID(c)

	// If user id does not match with token
	if !val {
		return
	}

	// Get teacher details
	teacher, err := app.models.Users.GetTeacherDetails(token.UserID)

	// Return error as internal server errror
	if err != nil {
		errBox.Add(data.InternalServerErrorResponse(err.Error()))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}

	// Return the teacher obj
	c.JSON(http.StatusOK, gin.H{"teacher": teacher})
}

// Show teacher handler
// Handler For GET "/v1/superusers/:user_id"
func (app *application) showSuperUserHandler(c *gin.Context) {
	// list of errors
	var errBox data.ErrorBox
	// Check if token matches with provided user ID
	val, token := app.DoesTokenMatchesUserID(c)

	// If user id does not match with token
	if !val {
		return
	}

	// Get superuser details
	su, err := app.models.Users.GetSuperUserDetails(token.UserID)

	// Return error as internal server errror
	if err != nil {
		errBox.Add(data.InternalServerErrorResponse(err.Error()))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}

	// Return the teacher obj
	c.JSON(http.StatusOK, gin.H{"superuser": su})
}

// Checks wether the tokens matches with the provided UserID
// That is, if a user is trying to access other's profile
func (app *application) DoesTokenMatchesUserID(c *gin.Context) (bool, *data.Token) {
	// list of errors
	var errBox data.ErrorBox
	// Get token value
	tokenVal := extractToken(c.GetHeader("Authorization"))

	userID := c.Param("user_id")

	// Incase user_id is empty
	if userID == "" {
		errBox.Add(data.BadRequestResponse("Invalid or missing user_id."))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return false, nil
	}

	userIDVal, err := strconv.Atoi(userID)

	// Conversion error
	if err != nil {
		errBox.Add(data.InternalServerErrorResponse("Please provide a valid user_id."))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)

		return false, nil
	}

	token, err := app.models.Tokens.GetTokenDetails(tokenVal)

	// incase of error
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			errBox.Add(data.AuthorizationErrorResponse("Invalid or expired token."))
			app.ErrorResponse(c, http.StatusUnauthorized, errBox)
			return false, nil
		default:
			errBox.Add(data.InternalServerErrorResponse(err.Error()))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)

			return false, nil
		}
	}

	// Incase the userID provided and the one on db donot match. Then 403 forbidden
	if token.UserID != int64(userIDVal) {
		errBox.Add(data.AuthorizationErrorResponse("You do not have authorization to access this resource."))
		app.ErrorResponse(c, http.StatusForbidden, errBox)
		return false, nil
	}

	// Return true
	return true, token
}
