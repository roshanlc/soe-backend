package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/roshanlc/soe-backend/internal/data"
)

// Handler for GET /v1/days
func (app *application) listDaysHandler(c *gin.Context) {

	// list of errors
	var errBox data.ErrorBox

	days, err := app.models.Schedule.GetAllDays()

	if err != nil {
		switch err {
		case data.ErrNoRecords:
			c.JSON(http.StatusOK, gin.H{"days": nil})
			return
		default:
			log.Println(err)
			errBox.Add(data.InternalServerErrorResponse("The server had some problems while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"days": days})
}

// Handler for GET "/v1/intervals"
func (app *application) listIntervalsHandler(c *gin.Context) {

	// list of errors
	var errBox data.ErrorBox

	intervals, err := app.models.Schedule.GetAllIntervals()

	if err != nil {
		switch err {
		case data.ErrNoRecords:
			c.JSON(http.StatusOK, gin.H{"intervals": nil})
			return
		default:
			log.Println(err)
			errBox.Add(data.InternalServerErrorResponse("The server had some problems while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"intervals": intervals})
}

// Handler for POST "/v1/schedules"
func (app *application) setScheduleHandler(c *gin.Context) {

	// list of errors
	var errBox data.ErrorBox

	var schedule data.Schedule

	// Bind the data
	err := c.ShouldBindJSON(&schedule)
	if err != nil {
		errBox.Add(data.BadRequestResponse(err.Error()))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}

	// Retrieve programID and semesterID
	programID := schedule.ProgramID
	semesterID := schedule.SemesterID

	_, err = app.models.Programs.GetProgram(programID)
	if err != nil {
		switch err {
		// 404 error
		case data.ErrRecordNotFound:
			errBox.Add(data.BadRequestResponse("The provided program_id does not exist."))
			app.ErrorResponse(c, http.StatusNotFound, errBox)
			return
		default:
			log.Println(err)
			errBox.Add(data.InternalServerErrorResponse("The server had some problems while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return

		}
	}

	_, err = app.models.Programs.GetSemester(semesterID)

	if err != nil {
		switch err {
		// 404 error
		case data.ErrRecordNotFound:
			errBox.Add(data.BadRequestResponse("The provided semester_id does not exist."))
			app.ErrorResponse(c, http.StatusNotFound, errBox)
			return
		default:
			log.Println(err)
			errBox.Add(data.InternalServerErrorResponse("The server had some problems while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}
	}

	err = app.models.Schedule.SetSchedule(&schedule)

	if err != nil {
		switch err {
		case data.ErrDuplicateEntry:
			errBox.Add(data.BadRequestResponse("Duplicate entry for schedule."))
			app.ErrorResponse(c, http.StatusBadRequest, errBox)
			return
		default:
			log.Println(err)
			errBox.Add(data.InternalServerErrorResponse("The server had some problems while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}

	}

	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Schedule Added", "The schedule for the semester was successfully added."))
	c.JSON(http.StatusCreated, gin.H{"messages": msgBox})

}

// Handler For GET "/v1/schedules/"
func (app *application) showScheduleHandler(c *gin.Context) {

	p, _ := strconv.Atoi((c.Query("program_id")))
	q, _ := strconv.Atoi(c.Query("semester_id"))
	x, _ := app.models.Schedule.GetSchedule(p, q)

	c.JSON(200, x)

}
