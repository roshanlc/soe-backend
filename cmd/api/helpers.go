package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/roshanlc/soe-backend/internal/data"
)

// Returns an error response back to user
func (app *application) ErrorResponse(c *gin.Context, code int, response interface{}) {

	// Set content-type header
	if c.Writer.Header().Get("Content-Type") != "application/json" {
		c.Writer.Header().Set("Content-Type", "application/json")
	}

	c.AbortWithStatusJSON(code, gin.H{"errors": response})
}

// Returns an error response for unallowed methods
func (app *application) MethodNotAllowedResponse(c *gin.Context) {
	var errBox data.ErrorBox
	errBox.Add(data.CustomErrorResponse("Method Not Allowed Error", c.Request.Method+" method is not allowed for this endpoint."))
	c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{"errors": errBox})
}

// Returns an error response for unimplemented routes
func (app *application) NoRouteResponse(c *gin.Context) {
	var errBox data.ErrorBox
	errBox.Add(data.ResourceNotFoundResponse(c.Request.URL.Path + " does not exist."))
	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"errors": errBox})
}
