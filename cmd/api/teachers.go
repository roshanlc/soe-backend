package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/roshanlc/soe-backend/internal/data"
)

// Show teacher handler
// Handler For GET "/v1/teachers/:user_id"
func (app *application) showTeacherHandler(c *gin.Context) {
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

// This update teacher Details
// Handler For POST "/v1/teachers/:user_id/update"
func (app *application) updateTeacherHandler(c *gin.Context) {
	// list of errors
	// var errBox data.ErrorBox
	// Check if token matches with provided user ID
	val, _ := app.DoesTokenMatchesUserID(c)

	// If user id does not match with token
	if !val {
		return
	}
	// TODO: add logic
}

func (app *application) listTeacherIssuesHandler(c *gin.Context) {

	var errBox data.ErrorBox

	// Check if token matches with provided user ID
	val, _ := app.DoesTokenMatchesUserID(c)

	// If user id does not match with token
	if !val {
		return
	}

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

	issues, err := app.models.Issues.GetIssues(userIDVal)

	if err != nil {
		switch err {
		case data.ErrNoRecords:
			c.JSON(http.StatusOK, gin.H{"issues": nil})
			return

		default:
			log.Println(err)
			errBox.Add(data.InternalServerErrorResponse("The server had problems while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}
	}

	c.JSON(http.StatusOK, issues)

}
