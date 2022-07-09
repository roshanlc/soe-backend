package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/roshanlc/soe-backend/internal/data"
)

type InputProfile struct {
	Profile           string   `json:"profile_id"`
	Experiences       []string `json:"experiences"`
	Publications      []string `json:"publications"`
	ResearchInterests []string `json:"research_interests"`
	Description       string   `json:"description"`
}

// listProfilesHandler returns list of teacers
// Handler for GET "/v1/profiles"
func (app *application) listProfilesHandler(c *gin.Context) {

	// slice of errors
	var errBox data.ErrorBox

	profiles, err := app.models.Profiles.GetAllPublicProfiles()

	if err != nil {
		log.Println(err)

		switch err {
		// empty records
		case data.ErrNoRecords:
			c.JSON(http.StatusOK, gin.H{"profiles": nil})
			return
		default:
			errBox.Add(data.InternalServerErrorResponse("The server had some problems while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"profiles": profiles})
}

// showProfileHandler returns a specific profile of a teacher
// Handler for GET "/v1/profiles/:profile_id"
func (app *application) showProfileHandler(c *gin.Context) {

	// slice of errors
	var errBox data.ErrorBox

	profileID := c.Param("profile_id")

	profile, err := app.models.Profiles.GetAPublicProfile(profileID)

	if err != nil {
		switch err {
		// 404 error
		case data.ErrRecordNotFound:
			errBox.Add(data.ResourceNotFoundResponse("No record found."))
			app.ErrorResponse(c, http.StatusNotFound, errBox)
			return
		default:
			log.Println(err)
			errBox.Add(data.InternalServerErrorResponse("The server had some problems while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return

		}
	}

	// Return the profile
	c.JSON(http.StatusOK, gin.H{"profile": profile})

}
