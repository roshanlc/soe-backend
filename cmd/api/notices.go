package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/roshanlc/soe-backend/internal/data"
)

/* listNoticesHandler returns the list of notices
from the database
*/
// listNoticesHandler godoc
// @Summary Lists notices
// @Description It retrieves all the notices from database
// @Tags notices
// @Produce json
// @Success 200 {array} data.Notice
// @Failure 400 {object} data.ErrorBox
// @Failure 500 {object} data.ErrorBox
// @Router /v1/notices [get]
func (app *application) listNoticesHandler(c *gin.Context) {

	// empty slcie containing all error messages
	var errArray data.ErrorBox

	limit, exists := c.GetQuery("limit")

	// Default limitVal is 20
	var limitVal int = 20

	// if query string is present
	if exists {

		limitVal, err := strconv.Atoi(limit)

		// Incase of parsing error
		if err != nil {

			errArray.Add(data.InternalServerErrorResponse("The server had a problem while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errArray)
			return

		} else if limitVal < 0 || limitVal > 1000 { // incase of bad limit value
			errArray.Add(data.BadRequestResponse("Provide a value for limit between 20 and 1000. Default is 20."))
			app.ErrorResponse(c, http.StatusBadRequest, errArray)
			return
		}

	}

	sort, exists := c.GetQuery("sort")

	// Default sort type is "desc"
	var sortVal string = "desc"

	// if query string is present
	if exists {
		sortVal = strings.ToLower(c.Query("sort"))

		// If unsupported sort type is provided
		if sort != "asc" && sort != "desc" {

			errArray.Add(data.BadRequestResponse("The provided sort type is not supported. Default is desc."))
			app.ErrorResponse(c, http.StatusBadRequest, errArray)
			return

		}
	}

	notices, err := app.models.Notices.GetAll(limitVal, sortVal)

	// If any error occured while retrieving records
	if err != nil {
		errArray.Add(data.InternalServerErrorResponse(err.Error()))
		app.ErrorResponse(c, http.StatusInternalServerError, errArray)
		return
	}

	// Return the notices
	c.JSON(http.StatusOK, gin.H{"notices": notices})
}

// Returns a specific notice from notice_id
func (app *application) showNoticeHandler(c *gin.Context) {

	// empty slcie containing all error messages
	var errArray data.ErrorBox

	id := c.Param("notice_id")

	idVal, err := strconv.Atoi(id)

	// Incase of error while parsing string into int type
	if err != nil {
		errArray.Add(data.InternalServerErrorResponse("The server had some problems while processing the request."))
		app.ErrorResponse(c, http.StatusInternalServerError, errArray)
		return
	} else if idVal < 0 {
		// incase of negative value of id
		errArray.Add(data.BadRequestResponse("Please provide a valid notice_id value."))
		app.ErrorResponse(c, http.StatusBadRequest, errArray)
		return
	}

	notice, err := app.models.Notices.Get(int64(idVal))

	// If some error has occured
	if err == data.ErrRecordNotFound {
		// incase of negative value of id
		errArray.Add(data.ResourceNotFoundResponse("The requested resource does not exist."))
		app.ErrorResponse(c, http.StatusNotFound, errArray)
		return
	} else if err != nil { // incase of any other errors
		errArray.Add(data.InternalServerErrorResponse("The server had some problems while processing the request."))
		app.ErrorResponse(c, http.StatusInternalServerError, errArray)
		return
	}

	// Return the notice
	c.JSON(http.StatusOK, gin.H{"notice": notice})
}
