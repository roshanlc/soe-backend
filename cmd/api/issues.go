package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/roshanlc/soe-backend/internal/data"
)

// listIssuesHandler returns the list of issues
// Handler for "GET" /v1/issues?filter=read/unread
func (app *application) listIssuesHandler(c *gin.Context) {

	var errBox data.ErrorBox

	filter := c.Query("filter")

	all, onlyRead := true, false

	if strings.EqualFold(filter, "read") {
		all = false
		onlyRead = true
	} else if strings.EqualFold(filter, "unread") {
		all = false
		onlyRead = false
	}

	issues, err := app.models.Issues.GetAllIssues(all, onlyRead)

	if err != nil {
		switch err {
		// Currently no records in the store
		case data.ErrNoRecords:
			c.JSON(http.StatusOK, gin.H{"issues": issues})
			return
		default:
			log.Println(err)
			errBox.Add(data.InternalServerErrorResponse("The server had problem while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}
	}

	// Return the issues
	c.JSON(http.StatusOK, gin.H{"issues": issues})
}

// registerIssueHandler creates a new issue
// Handler for "POST" /v1/issues
func (app *application) registerIssueHandler(c *gin.Context) {

	var errBox data.ErrorBox

	issue := c.PostForm("issue")

	token := extractToken(c.GetHeader("Authorization"))

	// If issue field is empty
	if issue == "" {

		errBox.Add(data.BadRequestResponse("issue field must not be empty."))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	// If issue characters are more than 1000 characters
	if len(issue) > 1000 {
		errBox.Add(data.BadRequestResponse("The text must not be more than 1000 characters"))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	err := app.models.Issues.RegisterIssue(issue, token)

	if err != nil {

		log.Println(err)
		errBox.Add(data.InternalServerErrorResponse("The server had problem while processing the request."))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}

	// success message
	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Issue Registered", "The issue was registered successfully."))

	// send success msg
	c.JSON(http.StatusCreated, gin.H{"messages": msgBox})
}

// Marks an issue as read
// Handler For GET "/v1/issues/:issue_id"
func (app *application) markIssueAsReadHandler(c *gin.Context) {
	// list of errors
	var errBox data.ErrorBox

	issueIDVal := c.Param("issue_id")

	isseID, err := strconv.Atoi(issueIDVal)

	if err != nil {
		log.Println(err)
		errBox.Add(data.InternalServerErrorResponse("The server had problem while processing the request."))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}

	if isseID <= 0 {
		errBox.Add(data.BadRequestResponse("The issue_id must be non-zero and non-negative value."))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}
	err = app.models.Issues.MarkAsRead(isseID)

	if err != nil {
		switch err {
		// No record found means 404 error
		case data.ErrRecordNotFound:
			errBox.Add(data.ResourceNotFoundResponse("The issue_id does not exist."))
			app.ErrorResponse(c, http.StatusNotFound, errBox)
			return

		default:
			log.Println(err)
			errBox.Add(data.InternalServerErrorResponse("The server had problem while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}
	}

	// success message
	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Issue Marked As Read", "The issue was marked as read successfully."))

	// send success msg
	c.JSON(http.StatusOK, gin.H{"messages": msgBox})
}
