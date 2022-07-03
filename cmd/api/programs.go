package main

import (
	"errors"
	"net/http"
	"strconv"

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

// Handler for GET /v1/semesters/
// Returns the list of semesters
func (app *application) listSemestersHandler(c *gin.Context) {
	// slice containing errors
	var errArray data.ErrorBox

	semesters, err := app.models.Programs.GetAllSemesters()

	if err != nil {
		switch {
		// Empty records
		case errors.Is(err, data.ErrNoRecords):
			c.JSON(http.StatusOK, gin.H{"semesters": nil})
			return
		default:
			errArray.Add(data.InternalServerErrorResponse(err.Error()))
			app.ErrorResponse(c, http.StatusInternalServerError, gin.H{"errors": errArray})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"semesters": semesters})
}

// Handler for GET "/v1/semesters/running?program_id="
func (app *application) listRunningSemestersHandler(c *gin.Context) {

	// slice containing errors
	var errBox data.ErrorBox

	p, exists := c.GetQuery("program_id")

	if !exists {
		errBox.Add(data.BadRequestResponse("Please provide program_id value."))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	programID, err := strconv.Atoi(p)
	if err != nil || programID <= 0 {
		errBox.Add(data.InternalServerErrorResponse("The server had problems while processing the request."))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}

	semesters, err := app.models.Programs.GetRunningSemesters(programID)

	if err != nil {
		switch err {
		// Empty records
		case data.ErrNoRecords:

			errBox.Add(data.BadRequestResponse("Please provide valid program_id."))
			app.ErrorResponse(c, http.StatusNotFound, errBox)
			return

		default:
			errBox.Add(data.InternalServerErrorResponse(err.Error()))
			app.ErrorResponse(c, http.StatusInternalServerError, gin.H{"errors": errBox})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"semesters": semesters})

}

// addRunningSemesterHandler adds a running semester to a program
func (app *application) addRunningSemesterHandler(c *gin.Context) {

	// slice containing errors
	var errBox data.ErrorBox

	var input data.Schedule

	err := c.ShouldBindJSON(&input)

	if err != nil {
		errBox.Add(data.BadRequestResponse(err.Error()))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	if input.ProgramID <= 0 || input.SemesterID <= 0 {
		errBox.Add(data.BadRequestResponse("Please provide valid program_id or semester_id value."))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	err = app.models.Programs.AddRunningSemester(input.ProgramID, input.SemesterID)

	if err != nil {
		switch err {
		case data.ErrDuplicateEntry:
			errBox.Add(data.BadRequestResponse("Duplicate entry."))
			app.ErrorResponse(c, http.StatusConflict, errBox)
			return
		default:
			errBox.Add(data.BadRequestResponse("Invalid program_id or semester_id."))
			app.ErrorResponse(c, http.StatusBadRequest, gin.H{"errors": errBox})
			return
		}
	}

	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Semester Added", "The semester was added successfully added as a running semester."))
	c.JSON(http.StatusCreated, gin.H{"messages": msgBox})
}
