package data

// A struct to hold configurations of the application
// New properties can be added easily with this.
type Config struct {
	Port int    // the port to run the backend at
	Env  string // other environment type (debug|release|test)

	DB struct { // database config

		Dsn          string // data source name
		MaxOpenConns int    // max number of connections that can be opened with db ( in-use + idle)
		MaxIdleConns int    // max idle connections to db
		MaxIdleTime  string // max idle time for a conn
	}
}
