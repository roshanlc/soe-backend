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

// struct to read passwords
type Pass struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// This retrieves user's basic details
// Handler for GET "/v1/users/:user_id"
func (app *application) showUserHandler(c *gin.Context) {

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

// Handler for POST "/v1/users/:user_id/password"
func (app *application) changePasswordHandler(c *gin.Context) {
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

	userID := c.Param("user_id")
	userIDVal, err := strconv.Atoi(userID)

	// Conversion error
	if err != nil {
		errBox.Add(data.InternalServerErrorResponse("Please provide a valid user_id."))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}
	var p Pass
	err = c.ShouldBindJSON(&p)

	if err != nil {
		errBox.Add(data.InternalServerErrorResponse("The server had problems when processing this request."))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}

	oldHash, err := app.models.Users.GetPassword(int64(userIDVal))

	if err != nil {
		switch err {
		case data.ErrRecordNotFound:
			errBox.Add(data.BadRequestResponse("Please provide a valid user_id."))
			app.ErrorResponse(c, http.StatusBadRequest, errBox)
			return
		default:
			errBox.Add(data.InternalServerErrorResponse("The server had problems when processing this request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return

		}
	}

	oldPass := data.Password{}
	oldPass.SetHash(oldHash)

	check, err := oldPass.Matches(p.OldPassword)
	if err != nil {
		errBox.Add(data.InternalServerErrorResponse("The server had problems when processing this request."))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}

	// If old password does not match
	if !check {

		errBox.Add(data.CustomErrorResponse("Password Mis-match", "The provided old password does not match."))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}
	// new password hash
	newPass := data.Password{}
	newPass.Set(p.NewPassword)

	err = app.models.Users.ChangePassword(int64(userIDVal), newPass.Hash())

	if err != nil {

		switch {
		case errors.Is(err, data.ErrNotUpdated):
			errBox.Add(data.CustomErrorResponse("Password Not Updated", "The password could not be updated"))
			app.ErrorResponse(c, http.StatusBadRequest, errBox)
			return
		default:
			errBox.Add(data.InternalServerErrorResponse("The server had problems when processing this request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}
	}

	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Password Changed", "The password was changed successfully."))
	c.JSON(http.StatusOK, gin.H{"messages": msgBox})
}
