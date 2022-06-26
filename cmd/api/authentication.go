package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/roshanlc/soe-backend/internal/data"
	"github.com/roshanlc/soe-backend/internal/validator"
)

// Handler for POST /v1/login endpoint
func (app *application) loginHandler(c *gin.Context) {

	// struct to hold input details
	var loginDetails data.LoginInput

	// empty slice to contain errors
	var errBox data.ErrorBox

	err := c.ShouldBindJSON(&loginDetails)

	// Incase of error
	if err != nil {

		errBox.Add(data.BadRequestResponse(err.Error()))

		app.ErrorResponse(c, http.StatusBadRequest, errBox)

		return
	}

	// Validate if email is proper and password length is greater than 8 characters

	v := validator.New()

	data.ValidateEmail(v, loginDetails.Email)

	// If errors
	if !v.Valid() {
		errBox.Add(data.CustomErrorResponse("Invalid Email", "Please provide a valid email address."))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	// We will check one by one, we will not make user aware about their mistakes all at once

	// Validate password if requirements are ok
	data.ValidatePasswordPlaintext(v, loginDetails.Password)

	if !v.Valid() {
		errBox.Add(data.CustomErrorResponse("Invalid Password", v.KeyValuePair("password")))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	// Check if email exists in database

	user, err := app.models.Users.GetByEmail(loginDetails.Email)

	if err != nil {
		switch {
		// No such email exists in database
		case errors.Is(err, data.ErrRecordNotFound):
			errBox.Add(data.InvalidCredentialsResponse("Please provide valid authentication details."))
			app.ErrorResponse(c, http.StatusUnauthorized, errBox)
			return

		// Incase of other errors
		default:
			errBox.Add(data.InternalServerErrorResponse("The server had problems while processing this request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}
	}

	// Now, validate if password matches
	match, err := user.Password.Matches(loginDetails.Password)

	if err != nil {

		errBox.Add(data.InternalServerErrorResponse(err.Error()))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}

	// If password does not match
	if !match {
		errBox.Add(data.InvalidCredentialsResponse("Please provide valid authentication details."))
		app.ErrorResponse(c, http.StatusUnauthorized, errBox)
		return
	}

	// If user is not activated
	if !user.Activated {
		errBox.Add(data.AccountErrorResponse("This account is yet to be activated."))
		app.ErrorResponse(c, http.StatusForbidden, errBox)
		return
	}

	// If user is expired
	if user.Expired {
		errBox.Add(data.AccountErrorResponse("This account has expired."))
		app.ErrorResponse(c, http.StatusForbidden, errBox)
		return
	}

	// Since password is correct
	// Delete any old tokens and
	// Generate new token and set expiry to 24 hrs from now

	// delete old tokens

	err = app.models.Tokens.DeleteAllForUser(data.ScopeAuthentication, user.UserID)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecords):
			// do nothing
		default:
			errBox.Add(data.InternalServerErrorResponse(err.Error()))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}

	}

	// now generae toke and insert into db

	token, err := app.models.Tokens.NewAuthenticationToken(user.UserID, time.Hour*24, data.ScopeAuthentication)

	if err != nil {
		errBox.Add(data.InternalServerErrorResponse(err.Error()))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}

	// Retrieve the role of user

	userRole, err := app.models.Roles.GetUserRole(user.UserID)

	if err != nil {
		errBox.Add(data.InternalServerErrorResponse(err.Error()))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}

	var authenicated data.Authentication = data.Authentication{UserID: token.UserID,
		Token:  token.Hash, // The token hash is to be returned
		Role:   userRole.Role.Name,
		Expiry: token.Expiry}

	// Return the authenticated details
	c.JSON(http.StatusOK, gin.H{"authentication": authenicated})
}

// Remove the authentication token for a user
// Handler For POST "/v1/logout"
func (app *application) logoutHandler(c *gin.Context) {

	var errBox data.ErrorBox
	tokenVal := extractToken(c.GetHeader("Authorization"))
	err := app.models.Tokens.LogoutUser(tokenVal)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			errBox.Add(data.BadRequestResponse("Please provide a valid token value in the Authorization header."))

		default:
			errBox.Add(data.InternalServerErrorResponse(err.Error()))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}

	}

	// Return successfull response
	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Logout Successfull", "The logout operation was performed successfully."))

	c.JSON(http.StatusOK, gin.H{"messages": msgBox})

}

// extract token from Authorization header
func extractToken(value string) string {
	token := ""
	// Authorization: Bearer TOKEN_VALUE_HERE

	headerParts := strings.Split(value, " ")

	// Token is the last CAPITAL ALPHANUMERIC STRING
	token = headerParts[1]

	return token
}
