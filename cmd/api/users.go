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
	// Get token value
	tokenVal := extractToken(c.GetHeader("Authorization"))

	userID := c.Param("user_id")

	// Incase user_id is empty
	if userID == "" {
		errBox.Add(data.BadRequestResponse("Invalid or missing user_id."))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	userIDVal, err := strconv.Atoi(userID)

	// Conversion error
	if err != nil {
		errBox.Add(data.InternalServerErrorResponse("Please provide a valid user_id."))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)

		return
	}

	token, err := app.models.Tokens.GetTokenDetails(tokenVal)

	// incase of error
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			errBox.Add(data.AuthorizationErrorResponse("Invalid or expired token."))
			app.ErrorResponse(c, http.StatusUnauthorized, errBox)
			return
		default:
			errBox.Add(data.InternalServerErrorResponse(err.Error()))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)

			return
		}
	}

	// Incase the userID provided and the one on db donot match. Then 403 forbidden
	if token.UserID != int64(userIDVal) {
		errBox.Add(data.AuthorizationErrorResponse("You do not have authorization to access this resource."))
		app.ErrorResponse(c, http.StatusForbidden, errBox)
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
	c.JSON(http.StatusOK, user)

}

// This retrieves student Details
// Handler For GET "/v1/students/:user_id"
func (app *application) showStudentHandler(c *gin.Context) {
	// list of errors
	var errBox data.ErrorBox
	// Get token value
	tokenVal := extractToken(c.GetHeader("Authorization"))

	userID := c.Param("user_id")

	// Incase user_id is empty
	if userID == "" {
		errBox.Add(data.BadRequestResponse("Invalid or missing user_id."))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	userIDVal, err := strconv.Atoi(userID)

	// Conversion error
	if err != nil {
		errBox.Add(data.InternalServerErrorResponse("Please provide a valid user_id."))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)

		return
	}

	token, err := app.models.Tokens.GetTokenDetails(tokenVal)

	// incase of error
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			errBox.Add(data.AuthorizationErrorResponse("Invalid or expired token."))
			app.ErrorResponse(c, http.StatusUnauthorized, errBox)
			return
		default:
			errBox.Add(data.InternalServerErrorResponse(err.Error()))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)

			return
		}
	}

	// Incase the userID provided and the one on db donot match. Then 403 forbidden
	if token.UserID != int64(userIDVal) {
		errBox.Add(data.AuthorizationErrorResponse("You do not have authorization to access this resource."))
		app.ErrorResponse(c, http.StatusForbidden, errBox)
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
	c.JSON(http.StatusOK, student)
}

// Show teacher handler
