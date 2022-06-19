package main

import (
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (app *application) serve() error {

	// Set the mode for server
	gin.SetMode(determineMode(app.config.Env))

	// A new empty gin.Engine
	router := gin.New()

	// Using the defaul logger and recovery middleware
	// They are quite good.
	// Todo: Change their config later
	router.Use(gin.Logger(), gin.Recovery())

	// Return proper response when unallowed method is sent
	router.HandleMethodNotAllowed = true

	// This handles the unallowed method requests
	router.NoMethod(app.MethodNotAllowedResponse)

	// Handles no routes requests
	router.NoRoute(app.NoRouteResponse)

	// Redirect fixed path i.e redirect /Foo  or /foo/ or /..//Foo to /foo
	router.RedirectFixedPath = true

	// Redirect  Trailing Slash i.e /foo/ will be redirected to /foo if only /foo route exists
	router.RedirectTrailingSlash = true

	// The max size for multi part memory
	router.MaxMultipartMemory = 10 << 20 // 10 MiB

	// Serve static content
	// By checking if the trimmed wildcard path *file is empty, you prevent queries to, e.g. /media/1/ (with final slash)
	// to list the directory. Instead /media/1 (without final slash) doesn't match any route
	// (it should automatically redirect to /media/1/).
	// Reference: https://stackoverflow.com/questions/69049626/how-to-serve-files-from-dynamic-subdirectories-using-gin
	router.GET("/uploads/notices/:folder/*file", func(c *gin.Context) {
		folder := c.Param("folder")
		file := c.Param("file")
		if strings.TrimPrefix(file, "/") == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		fullName := filepath.Join("uploads/notices", folder, filepath.FromSlash(path.Clean("/"+file)))

		// Set the cache for 7 days
		c.Header("Cache-Control", "public, max-age=604800")
		// return the file
		c.File(fullName)
	})

	// Simple group: v1
	v1 := router.Group("/v1")
	{
		// notices
		v1.GET("/notices", app.listNoticesHandler)
		v1.GET("/notices/:notice_id", app.showNoticeHandler)

		// Requires superuser authorization for publication and deletion of notices
		v1.POST("/notices", app.limitUploadSize, app.isAdmin, app.publishNoticeHandler)
		v1.DELETE("/notices/:notice_id", app.isAdmin, app.deleteNoticeHandler)

		// courses
		v1.GET("/courses", app.listCoursesHandler)
		v1.GET("/courses/:course_code", app.showCourseHandler)

		// teacher profiles
		v1.GET("/profiles", app.listProfilesHandler)
		v1.GET("/profiles/:profie_name", app.showProfileHandler)

		// authentication handler
		v1.POST("/login", app.limitBodySize, app.loginHandler)       // login operation
		v1.POST("/logout", app.authenticatedUser, app.logoutHandler) // logout operation

		// user details handler

		// First checks if user is logged in, then only passes to final stage
		v1.GET("/users/:user_id", app.authenticatedUser, app.showUserHandler)

		v1.GET("/students/:user_id", app.authenticatedUser, app.showStudentHandler)
		v1.GET("/teachers/:user_id", app.authenticatedUser, app.showTeacherHandler)
		v1.GET("/superusers/:user_id", app.authenticatedUser, app.showSuperUserHandler)

		// Register users
		v1.POST("/students/register", app.registerStudentHandler) // Register a student
		v1.POST("/teachers/register", app.registerTeacherHandler) // Register a teacher

		// Programs and levels
		v1.GET("/programs", app.listProgramsHandler)
		v1.GET("/levels", app.listLevelsHandler)

	}

	// server struct
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil

}

func determineMode(mode string) string {

	switch mode {
	case "debug":
		return gin.DebugMode
	case "release":
		return gin.ReleaseMode
	case "test":
		return gin.TestMode

	default:
		return gin.DebugMode
	}
}
