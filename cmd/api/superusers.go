package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/roshanlc/soe-backend/internal/data"
)

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
