package main

import (
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
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

	cor := cors.DefaultConfig()
	cor.AllowOrigins = app.config.CORS

	// Setup CORS policy
	router.Use(cors.New(cor))

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
		v1.GET("/profiles/:profile_id", app.showProfileHandler)
		v1.POST("/teachers/:user_id/profile", app.isTeacher, app.createProfileHandler)
		v1.PUT("/teachers/:user_id/profile", app.isTeacher, app.updateProfileHandler)
		v1.DELETE("/teachers/:user_id/profile", app.isTeacher, app.deleteProfileHandler)

		// authentication handler
		v1.POST("/login", app.limitBodySize, app.loginHandler)       // login operation
		v1.POST("/logout", app.authenticatedUser, app.logoutHandler) // logout operation

		// user details handler

		// First checks if user is logged in, then only passes to final stage
		v1.GET("/users/:user_id", app.authenticatedUser, app.showUserHandler)

		v1.GET("/students/:user_id", app.authenticatedUser, app.showStudentHandler)
		v1.GET("/teachers/:user_id", app.authenticatedUser, app.showTeacherHandler)
		v1.GET("/superusers/:user_id", app.authenticatedUser, app.showSuperUserHandler)

		// Change Password
		v1.POST("/users/:user_id/password", app.authenticatedUser, app.changePasswordHandler)

		// Update student's and teacher's details (business logic is yet to be added )
		v1.POST("/students/:user_id/update", app.authenticatedUser, app.updateStudentHandler)
		v1.POST("/teachers/:user_id/update", app.authenticatedUser, app.updateTeacherHandler)

		// Register users
		v1.POST("/students/register", app.registerStudentHandler) // Register a student
		v1.POST("/teachers/register", app.registerTeacherHandler) // Register a teacher

		// Activate users
		v1.GET("/users/activate", app.activateUserHandler)

		// Programs and levels
		v1.GET("/faculties", app.listFacultiesHandler)
		v1.GET("/faculties/:faculty_id", app.showFacultyHandler)

		v1.GET("/departments", app.listDepartmentsHandler)
		v1.GET("/departments/:faculty_id", app.showDepartmentsHandler)

		v1.GET("/programs", app.listProgramsHandler)
		v1.GET("/programs/:program_id", app.showProgramHandler)

		v1.GET("/levels", app.listLevelsHandler)
		v1.GET("/semesters", app.listSemestersHandler)
		v1.GET("/semesters/running", app.listRunningSemestersHandler)
		v1.POST("/semesters/running", app.isAdmin, app.addRunningSemesterHandler)

		// Schedules
		v1.GET("/days", app.listDaysHandler)
		v1.GET("/intervals", app.listIntervalsHandler)
		v1.POST("/schedules", app.isAdmin, app.setScheduleHandler)
		v1.DELETE("/schedules", app.isAdmin, app.deleteScheduleHandler)

		v1.GET("/schedules", app.showScheduleHandler)
		v1.GET("/teachers/:user_id/schedule", app.isTeacher, app.showTeacherScheduleHandler)
		v1.GET("/students/:user_id/schedule", app.isStudent, app.showStudentScheduleHandler)

		// Issues

		v1.GET("/issues", app.isAdmin, app.listIssuesHandler)
		v1.POST("/issues", app.isStudentOrTeacher, app.registerIssueHandler) // For students and teachers
		v1.GET("/students/:user_id/issues", app.isStudent, app.listStudentIssuesHandler)
		v1.GET("/teachers/:user_id/issues", app.isTeacher, app.listTeacherIssuesHandler)
		v1.PUT("/issues/:issue_id", app.isAdmin, app.markIssueAsReadHandler)

	}

	// server struct
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Ticker to remove expired tokens routinely
	ticker := time.NewTicker(1 * time.Minute)

	// Run a separate go routine
	go func() {

		// Range over the ticker
		for range ticker.C {

			// Call the token removing method
			app.expiredTokenRemoval()
		}

	}()

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
