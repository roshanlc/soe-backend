package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/roshanlc/soe-backend/internal/data"
)

// listCoursesHandler returns the list of courses
// Handler for GET "/v1/courses"
func (app *application) listCoursesHandler(c *gin.Context) {

	// slice containing errors
	var errArray data.ErrorBox

	courses, err := app.models.Courses.GetAll()

	if err != nil {
		app.logger.PrintError(err, nil)

		switch err {

		// Incase of 404 response
		case data.ErrRecordNotFound:

			errArray = append(errArray, data.ResourceNotFoundResponse("The requested resource does not exist."))
			app.ErrorResponse(c, http.StatusNotFound, errArray)
			return

		case data.ErrNoRecords:
			c.JSON(http.StatusOK, gin.H{"courses": nil})
			return

		// Internal server error response
		default:
			fmt.Println(err)

			errArray.Add(data.InternalServerErrorResponse("The server had some problems while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errArray)
			return
		}
	}

	// Return the course detail
	c.JSON(http.StatusOK, gin.H{"courses": courses})
}

// Returns a specific course details
// Handler for GET "/v1/courses/:course_code"
func (app *application) showCourseHandler(c *gin.Context) {

	// slice containing errors
	var errArray data.ErrorBox

	courseCode := c.Param("course_code")

	if courseCode == "" {

		// Add bad request error to errorBox
		errArray = append(errArray, data.BadRequestResponse("Please provide a valid course_code value."))
		app.ErrorResponse(c, http.StatusBadRequest, errArray)
		return
	}

	course, err := app.models.Courses.Get(courseCode)

	if err != nil {

		app.logger.PrintError(err, nil)

		switch err {

		// Incase of 404 response
		case data.ErrRecordNotFound:

			errArray = append(errArray, data.ResourceNotFoundResponse("The requested resource does not exist."))
			app.ErrorResponse(c, http.StatusNotFound, errArray)
			return

		// Internal server error response
		default:
			fmt.Println(err)

			errArray.Add(data.InternalServerErrorResponse("The server had some problems while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errArray)
			return
		}
	}

	// Return the course detail
	c.JSON(http.StatusOK, gin.H{"course": course})
}
