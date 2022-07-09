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

// createProfileHandler creates a public profile for a teacher
// Handler for POST "/v1/teacher/:user_id/profile"
func (app *application) createProfileHandler(c *gin.Context) {

	// list of errors
	var errBox data.ErrorBox

	// Check if token matches with provided user ID
	val, _ := app.DoesTokenMatchesUserID(c)

	// If user id does not match with token
	if !val {
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

	var profile InputProfile

	err = c.ShouldBindJSON(&profile)

	if err != nil {
		errBox.Add(data.BadRequestResponse("Malformed request body: " + err.Error()))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	err = app.models.Profiles.CreatePublicProfile(userIDVal,
		profile.Profile,
		profile.Experiences,
		profile.Publications,
		profile.ResearchInterests,
		profile.Description)

	// Incase of error
	if err != nil {
		switch err {
		case data.ErrDuplicateProfileID:
			errBox.Add(data.CustomErrorResponse("Duplicate profile_id", "Please use a different profile_id value."))
			app.ErrorResponse(c, http.StatusConflict, errBox)
			return
		case data.ErrDuplicateEntry:
			errBox.Add(data.BadRequestResponse("Duplicate entry. Please use PUT method to update the profile details."))
			app.ErrorResponse(c, http.StatusBadRequest, errBox)
			return

		default:
			log.Println(err)
			errBox.Add(data.InternalServerErrorResponse("The server had some problems while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}
	}

	// success message
	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Profile Created", "The profile was created successfully."))

	// send success msg
	c.JSON(http.StatusCreated, gin.H{"messages": msgBox})

}

// updateProfileHandler updates the public profile of a teacher
// Handler for PUT "/v1/teachers/:user_id/profile"
func (app *application) updateProfileHandler(c *gin.Context) {

	// list of errors
	var errBox data.ErrorBox

	// Check if token matches with provided user ID
	val, _ := app.DoesTokenMatchesUserID(c)

	// If user id does not match with token
	if !val {
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
	var profile InputProfile

	err = c.ShouldBindJSON(&profile)

	if err != nil {
		errBox.Add(data.BadRequestResponse("Malformed request body: " + err.Error()))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	err = app.models.Profiles.UpdatePublicProfile(userIDVal, profile.Profile, profile.Experiences, profile.Publications, profile.ResearchInterests, profile.Description)

	if err != nil {
		switch err {
		// Trying to update a non-existing record
		case data.ErrRecordNotFound:
			errBox.Add(data.ResourceNotFoundResponse("Record does not exist"))
			app.ErrorResponse(c, http.StatusNotFound, errBox)
			return

		// Certain problems while updating record
		default:
			errBox.Add(data.InternalServerErrorResponse("The server had problems while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}
	}

	// success message
	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Profile Updated", "The profile was updated successfully."))

	// send success msg
	c.JSON(http.StatusOK, gin.H{"messages": msgBox})
}

// deleteProfileHandler deletes the public profile of a teacher
// Handler for DELETE "/v1/teachers/:user_id/profile"
func (app *application) deleteProfileHandler(c *gin.Context) {

	// list of errors
	var errBox data.ErrorBox

	// Check if token matches with provided user ID
	val, _ := app.DoesTokenMatchesUserID(c)

	// If user id does not match with token
	if !val {
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

	err = app.models.Profiles.DeleteProfile(userIDVal)

	if err != nil {
		switch err {
		// Trying to update a non-existing record
		case data.ErrRecordNotFound:
			errBox.Add(data.ResourceNotFoundResponse("Record does not exist"))
			app.ErrorResponse(c, http.StatusNotFound, errBox)
			return

		// Certain problems while updating record
		default:
			errBox.Add(data.InternalServerErrorResponse("The server had problems while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}
	}

	// success message
	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Profile Deleted", "The profile was deleted successfully."))

	// send success msg
	c.JSON(http.StatusOK, gin.H{"messages": msgBox})
}
