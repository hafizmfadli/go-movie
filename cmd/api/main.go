package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"sync"
	"time"

	"github.com/hafizmfadli/go-movie/internal/data"
	"github.com/hafizmfadli/go-movie/internal/jsonlog"
	"github.com/hafizmfadli/go-movie/internal/mailer"
	_ "github.com/lib/pq"
)

// Application version number
const version = "1.0.0"

// config struct hold all the configuration settings for out application.
type config struct {

	// the network port that we want the server to listen on
	port int

	// current operating environment for the application (dev, staging, prod, etc..)
	env string

	// db struct field hold the configuration settings for our database connection pool.
	db struct  {
		dsn string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime string
	}

	// limiter struct containing fields for the requests per second and burst
	// values, and a boolean field which we can use to enable/disable rate limiting
	// altogether
	limiter struct {
		rps float64
		burst int
		enabled bool
	}

	// smtp struct hold smtp configuration
	smtp struct {
		host string
		port int
		username string
		password string
		sender string
	}
}

// application struct hold the dependencies for our HTTP handlers, helpers, and middleware.
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	// sync.WaitGroup is used to coordinate the graceful shutdown and our background goroutine
	wg sync.WaitGroup
}

func main(){

	var cfg config

	// Read the value of the port and enc command-line flags into the config struct.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("NETFLIX_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum request per seocnd")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.StringVar(&cfg.smtp.host, "smtp-host", "127.0.0.1", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 1025, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "hafiz", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "pa55word", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Netflix <no-reply@netflix.hafizmfadli.net>", "SMTP sender")

	flag.Parse()

	// Initialize a new jsonlog.Logger kwhich writes any messages *at or above* the INFO
	// severity level to the standard out stream
	logger := jsonlog.NewLogger(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

// openDB returns a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	// create an empty connection pool
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections in the pool.
	// Note that passing a value less than or equal to 0 will mean there is no limit.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set the maximum number of idle connections in the pool. Again, passing a value
	// less than or equal to 0 will mean there is no limit.
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set the maximum idle timeout
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	// establish a new connection to the database. If the connection couldn't be
	// established successfully within the 5 second deadline, then this will return an error
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}	