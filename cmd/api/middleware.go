package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/roshanlc/soe-backend/internal/data"
)

// This middleware limits the size of request Body to 1 MB. If request size exceeds,
// it returns a bad request response
func (app *application) limitBodySize(c *gin.Context) {

	// use http.MaxBytesReader to limit the size of the request body to the provided kilo Bytes
	// to prevent from request body overload attacks

	var maxBytes int64 = 1_048_576 // 1 MiB

	// Length Of Request (Bytes)
	length := c.Request.ContentLength

	// If request length has exceeded the limit
	if length > maxBytes {

		var errBox data.ErrorBox
		errBox.Add(data.CustomErrorResponse("Payload Too Large", "The request body is too large to process. The maximum request body size is 1 MB."))

		// Abort the request and return an error response
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"errors": errBox})
	}

	// Pass the request
	c.Next()
}

// This middleware limits the total size of upload in requests.
// If size exceeds it returns a bad request.
func (app *application) limitUploadSize(c *gin.Context) {

	// use http.MaxBytesReader to limit the size of the request body to the provided kilo Bytes
	// to prevent from request body overload attacks

	var maxBytes int64 = 50_048_576 // 50 MiB

	// Length Of Request (Bytes)
	length := c.Request.ContentLength

	// If request length has exceeded the limit
	if length > maxBytes {

		var errBox data.ErrorBox
		errBox.Add(data.CustomErrorResponse("Payload Too Large", "The request body is too large to process. The maximum request body size is 50 MB."))

		// Abort the request and return an error response
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"errors": errBox})
	}

	// Pass the request
	c.Next()
}

// Checks if the user is authenticated or not
func (app *application) authenticatedUser(c *gin.Context) {

	// Errors Box
	var errBox data.ErrorBox

	// Add the "Vary: Authorization" header to the response. This indicates to any
	// caches that the response may vary based on the value of the Authorization
	// header in the request.
	c.Set("Vary", "Authorization")

	// Retrieve the value of the Authorization header from the request. This will
	// return the empty string "" if there is no such header found.
	authHeader := c.GetHeader("Authorization")

	// If the header is missing
	if authHeader == "" {
		errBox.Add(data.BadRequestResponse("Please provide a token value in the Authorization header."))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": errBox})
		return
	}

	// Otherwise, we expect the value of the Authorization header to be in the format
	// "Bearer <token>". We try to split this into its constituent parts, and if the
	// header isn't in the expected format we return a 400 Bad Request response
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		errBox.Add(data.BadRequestResponse("Please provide a token value in the Authorization header."))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": errBox})
		return

	}

	// Extract the actual authentication token from the header parts.
	token := headerParts[1]

	// Check if token format is valid
	if !validTokenLength(token) {
		errBox.Add(data.BadRequestResponse("Please provide a token value in the Authorization header."))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": errBox})
		return
	}

	// Check if token is valid
	_, err := app.models.Tokens.LoggedIn(token)
	if err != nil {

		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			// invalid token
			errBox.Add(data.AuthorizationErrorResponse("Invalid or expired token."))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"errors": errBox})
			return
		default:
			errBox.Add(data.InternalServerErrorResponse("The server had problems while processing this request."))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"errors": errBox})
			return
		}
	}

	// Since token is valid, pass the request
	c.Next()
}

// Checks if the user is admin or not
func (app *application) isAdmin(c *gin.Context) {

	// Errors Box
	var errBox data.ErrorBox

	// Add the "Vary: Authorization" header to the response. This indicates to any
	// caches that the response may vary based on the value of the Authorization
	// header in the request.
	c.Set("Vary", "Authorization")

	// Retrieve the value of the Authorization header from the request. This will
	// return the empty string "" if there is no such header found.
	authHeader := c.GetHeader("Authorization")

	// If the header is missing
	if authHeader == "" {
		errBox.Add(data.BadRequestResponse("Please provide a token value in the Authorization header."))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": errBox})
		return
	}

	// Otherwise, we expect the value of the Authorization header to be in the format
	// "Bearer <token>". We try to split this into its constituent parts, and if the
	// header isn't in the expected format we return a 400 Bad Request response
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		errBox.Add(data.BadRequestResponse("Please provide a token value in the Authorization header."))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": errBox})
		return

	}

	// Extract the actual authentication token from the header parts.
	token := headerParts[1]

	// Check if token format is valid
	if !validTokenLength(token) {
		errBox.Add(data.BadRequestResponse("Please provide a token value in the Authorization header."))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": errBox})
		return
	}

	// Check if token is valid
	t, err := app.models.Tokens.LoggedIn(token)
	if err != nil {

		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			// invalid token
			errBox.Add(data.AuthorizationErrorResponse("Invalid or expired token."))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"errors": errBox})
			return
		default:
			errBox.Add(data.InternalServerErrorResponse("The server had problems while processing this request."))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"errors": errBox})
			return
		}
	}

	// Get user's role
	r, err := app.models.Roles.GetUserRole(t.UserID)

	if err != nil {

		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			// invalid token
			errBox.Add(data.AuthorizationErrorResponse("Invalid or expired token."))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"errors": errBox})
			return
		default:
			errBox.Add(data.InternalServerErrorResponse("The server had problems while processing this request."))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"errors": errBox})
			return
		}
	}

	// Check if the role is superuser or not
	if strings.ToLower(r.Role.Name) != "superuser" {
		errBox.Add(data.AuthorizationErrorResponse("You donot have authorization access this resource."))
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"errors": errBox})
		return
	}

	// Since token is valid, pass the request
	c.Next()
}

// validTokenLength checks if the provided token is of valid length
// A valid token is of upper cases and has length of 32 chars
func validTokenLength(token string) bool {

	// If length is 32 and if token is upper cased then, return true
	if len(token) == 32 && strings.ToUpper(token) == token {
		return true
	}
	return false
}
