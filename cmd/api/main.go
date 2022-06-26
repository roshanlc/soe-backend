package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/roshanlc/soe-backend/internal/data"
	"github.com/roshanlc/soe-backend/internal/jsonlog"
	"github.com/roshanlc/soe-backend/internal/utils"
	"golang.org/x/net/context"
)

// Application version
// Later, it will be generated automatically during build process
const version = "1.0.0"

// A struct to hold all the dependencies of the application
// such as handlers, middlewares, database connection and so on.
// It can grow as needed.
type application struct {
	config      *data.Config
	logger      *jsonlog.Logger
	models      data.Models
	mailHandler *MailingContainer
}

func main() {

	// Create a new version boolean flag with the default value of false.
	displayVersion := flag.Bool("version", false, "Display version and exit")
	displayHelp := flag.Bool("help", false, "Display help information.")

	// Parse the flags
	flag.Parse()

	// If the help version is true, then print out help info and
	// immediately exit.
	if *displayHelp {
		fmt.Println(displayHelpInfo())
		os.Exit(0)
	}

	// If the version flag value is true, then print out the version number and
	// immediately exit.
	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	cfg, err := utils.ReadConfig("config.toml")
	if err != nil {
		log.Fatal("Unable to read config.toml, ", err)
		return
	}

	// Initalize a new logger which writes on stdout
	// prefixed with current date and time
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	logger.PrintInfo("Config file has been loaded.", nil)

	// setupFolders
	setupFolders()

	// Call the openDB() helper function to create the connection pool,
	// passing in the config struct. It this error returns an error, we log it and exit
	// the application immediately

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// Defer a call to db.Close() so that the connection pool is closed before
	// the main() function terminates.
	defer db.Close()

	// Also log the message that db connection is successfull
	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config:      cfg,
		logger:      logger,
		models:      data.NewModels(db),
		mailHandler: NewMailer(),
	}

	// Start mailer
	err = app.mailHandler.Authenticate(app.config)

	if err != nil {
		app.logger.PrintFatal(err, nil)
	}

	err = app.serve()

	if err != nil {
		app.logger.PrintFatal(err, nil)
	}

}

// openDB returns a sql.DB connection pool
func openDB(config *data.Config) (*sql.DB, error) {

	db, err := sql.Open("postgres", config.DB.Dsn)

	if err != nil {
		return nil, err
	}

	// Set the database connection settings such as
	// max open conns, max idle time and max idle conns
	db.SetMaxOpenConns(config.DB.MaxOpenConns)
	db.SetMaxIdleConns(config.DB.MaxIdleConns)

	// Use time.ParseDuration() function to convert the idle timout duration string
	// to a time.Duration type
	duration, err := time.ParseDuration(config.DB.MaxIdleTime)

	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5 sec timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// defer cancel
	defer cancel()

	// Use PingContext() to establish a new connection to the database, passing
	// the context we just created above as a parameter. If the connection couldnot be
	// established within the 5 sec deadline, it will return an error
	err = db.PingContext(ctx)

	if err != nil {
		return nil, err
	}

	// Return the sql.DB connection pool
	return db, nil
}

// Returns help info
func displayHelpInfo() string {

	var helpInfo string = fmt.Sprintf("Version: %v\n", version)
	helpInfo += "The configuration should be in \"config.toml\".\nCheck the config_example.toml for sample configuration."
	return helpInfo
}

// setupFolders setups necessary folders
func setupFolders() {

	uploadsFolder := "./uploads/notices"

	// Create directories
	os.MkdirAll(uploadsFolder, 0777) // IDK why but 0777 is the only permission that allows creating new files/dirs inside it
}
