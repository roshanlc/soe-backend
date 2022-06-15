package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/roshanlc/soe-backend/internal/data"
)

// Handler for GET /v1/levels/
// Returns the list of levels
func (app *application) listLevelsHandler(c *gin.Context) {

	// slice containing errors
	var errArray data.ErrorBox

	levels, err := app.models.Programs.GetAllLevels()

	if err != nil {
		switch {
		// Empty records
		case errors.Is(err, data.ErrNoRecords):
			c.JSON(http.StatusOK, gin.H{"levels": nil})
			return
		default:
			errArray.Add(data.InternalServerErrorResponse(err.Error()))
			app.ErrorResponse(c, http.StatusInternalServerError, gin.H{"errors": errArray})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"levels": levels})
}

// Handler for GET /v1/programs/
// Returns the list of programs
func (app *application) listProgramsHandler(c *gin.Context) {
	// slice containing errors
	var errArray data.ErrorBox

	// Filters
	var level, department, faculty string

	level, _ = c.GetQuery("level")

	department, _ = c.GetQuery("department")

	faculty, _ = c.GetQuery("faculty")

	programs, err := app.models.Programs.GetAllPrograms(level, department, faculty)

	if err != nil {
		switch {
		// Empty records
		case errors.Is(err, data.ErrNoRecords):
			c.JSON(http.StatusOK, gin.H{"programs": nil})
			return
		default:
			errArray.Add(data.InternalServerErrorResponse(err.Error()))
			app.ErrorResponse(c, http.StatusInternalServerError, gin.H{"errors": errArray})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"programs": programs})
}
