package main

import (
	"github.com/gin-gonic/gin"
)

// listProfilesHandler returns list of teacers
// Handler for GET "/v1/profiles"
func (app *application) listProfilesHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Under construction."})
}

// showProfileHandler returns a specific profile of a teacher
// Handler for GET "/v1/profiles/:profile_name"
func (app *application) showProfileHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Under construction."})

}
