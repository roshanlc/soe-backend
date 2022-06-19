package main

import (
	"errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/roshanlc/soe-backend/internal/data"
)

// Relative path to folder where notices' media content will be stored
const PathToNoticesStorage = "./uploads/notices/" // Use this for actual storage
const PathToNoticesURL = "/uploads/notices/"      // Use this for url

// List of supported file mime types
var SupportedFileType = []string{"application/pdf", "image/png", "image/jpeg"}

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

// publishNoticeHandler publishes notices
// Handler for POST /v1/notices
func (app *application) publishNoticeHandler(c *gin.Context) {

	// box of errors
	var errBox data.ErrorBox

	// List of file path
	var filepaths []string

	// folder name
	// time format= 2022-06-18 13:50
	var foldername string = string(time.Now().Format("20060102T1504"))
	var folder string = PathToNoticesStorage + foldername

	err := os.Mkdir(folder, 0777)

	if err != nil {
		app.logger.PrintError(err, nil)
		errBox.Add(data.InternalServerErrorResponse("The server had a problem while processing the request."))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}

	maxSize := 10_048_576 // 10 MB

	form, err := c.MultipartForm()

	// If no errors
	if err != nil {
		app.logger.PrintError(err, nil)

		// delete the created folder
		deleteFolder(folder)
		errBox.Add(data.InternalServerErrorResponse("The server had a problem while processing the request."))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return
	}
	var title, content string

	// title of notice
	if len(form.Value["title"]) == 0 {
		// delete the created folder
		deleteFolder(folder)
		errBox.Add(data.BadRequestResponse("Title missing."))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return

	}

	if len(form.Value["content"]) == 0 {
		// delete the created folder
		deleteFolder(folder)
		errBox.Add(data.BadRequestResponse("Content missing."))
		app.ErrorResponse(c, http.StatusBadRequest, errBox)
		return
	}
	title = form.Value["title"][0]
	content = form.Value["content"][0]

	files := form.File["media"]

	// Check the total number of files uploaded
	if len(files) > 10 {
		// delete the created folder
		deleteFolder(folder)
		errBox.Add(data.CustomErrorResponse("Too Many Payload Files", "The maximum number of files that can uploaded is 10."))
		app.ErrorResponse(c, http.StatusRequestEntityTooLarge, errBox)
		return
	}

	// Check if a file exceeds 10MB or is unsupported
	for _, file := range files {

		// If a file exceeds 10 MB size
		if file.Size > int64(maxSize) {
			// delete the created folder
			deleteFolder(folder)
			errBox.Add(data.CustomErrorResponse("Too Many Payload Files", "The maximum number of files that can uploaded is 10."))
			app.ErrorResponse(c, http.StatusRequestEntityTooLarge, errBox)
			return
		}

		right, err := validContentType(file)

		if err != nil {
			app.logger.PrintError(err, nil)

			// delete the created folder
			deleteFolder(folder)
			errBox.Add(data.InternalServerErrorResponse("The server had a problem while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}

		// Incase of invalid content type
		if !right {
			// delete the created folder
			deleteFolder(folder)
			errBox.Add(data.CustomErrorResponse("Unsupported Media Type", "The supported media types are pdf, jpeg/jpg and png."))
			app.ErrorResponse(c, http.StatusUnsupportedMediaType, errBox)
			return
		}

		filePath, _ := filepath.Abs(folder + "/" + file.Filename)

		// use for url
		fileUrl := PathToNoticesURL + foldername + "/" + file.Filename

		// Save the file
		err = saveFile(file, filePath)

		if err != nil {
			// Delete the folder
			app.logger.PrintError(err, nil)
			deleteFolder(folder)
			errBox.Add(data.InternalServerErrorResponse("The server had a problem while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errBox)
			return
		}
		// Save a list of files
		filepaths = append(filepaths, fileUrl)
	}

	// Get the token value
	tokenVal := extractToken(c.GetHeader("Authorization"))

	// Now insert into db
	err = app.models.Notices.Insert(title, content, filepaths, tokenVal)

	if err != nil {
		app.logger.PrintError(err, nil)

		// delete the created folder
		deleteFolder(folder)
		errBox.Add(data.InternalServerErrorResponse("The server had a problem while processing the request."))
		app.ErrorResponse(c, http.StatusInternalServerError, errBox)
		return

	}

	// success message
	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Notice Created", "The notice was created successfully."))

	// send success msg
	c.JSON(http.StatusCreated, gin.H{"messages": msgBox})

}

// Handler for DELETE /v1/notices/:notice_id
func (app *application) deleteNoticeHandler(c *gin.Context) {

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

	err = app.models.Notices.Delete(int64(idVal))

	if err != nil {

		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			// incase of negative value of id
			errArray.Add(data.ResourceNotFoundResponse("Please provide a valid notice_id value."))
			app.ErrorResponse(c, http.StatusNotFound, errArray)
			return

		default:
			errArray.Add(data.InternalServerErrorResponse("The server had some problems while processing the request."))
			app.ErrorResponse(c, http.StatusInternalServerError, errArray)
			return
		}
	}

	// message box
	var msgBox data.MessageBox
	msgBox.Add(data.MessageResponse("Notice Deleted",
		"The notice was deleted successfully."))

	// return success message
	c.JSON(http.StatusOK, gin.H{"messages": msgBox})
}

// validContentType checks if a file is of supported type or not
func validContentType(f *multipart.FileHeader) (bool, error) {
	// Open the file to check content type
	d, err := f.Open()

	// If  errors
	if err != nil {
		return false, err
	}

	defer d.Close()
	a, err := ioutil.ReadAll(d)

	// If  errors
	if err != nil {
		return false, err
	}

	t := http.DetectContentType(a)

	// If content-type matches with saved ones
	for _, val := range SupportedFileType {
		if strings.ToLower(t) == val {
			return true, nil
		}
	}
	return false, nil
}

// Save file
func saveFile(f *multipart.FileHeader, filename string) error {

	file, err := os.Create(filename)
	// If  errors
	if err != nil {
		return err
	}

	defer file.Close()

	// Open the file
	d, err := f.Open()

	// If  errors
	if err != nil {
		return err
	}
	defer d.Close()
	a, err := ioutil.ReadAll(d)

	// If  errors
	if err != nil {
		return err
	}

	_, err = file.Write(a)
	if err != nil {
		return err
	}

	return nil
}

// deleteFolder deletes a folder
func deleteFolder(folder string) {

	os.RemoveAll(folder)
}
