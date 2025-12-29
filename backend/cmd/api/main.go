package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"khayal/internal/data"
	"khayal/internal/jsonlog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// NOTE: Write good comments

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func main() {

	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "environment(dev|staging|prod)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://khayal:pa55word@localhost/khayal", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "maximum number of open connections to the database")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "maximum number of idle connections to the database")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "maximum amount of time a connection may be idle")

	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	logger.PrintInfo("database connection established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}
	// overriding the default for these, there are many methods for this struct
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(), // handles the requests
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  cfg.env,
	})

	err = srv.ListenAndServe() // initializes
	logger.PrintFatal(err, nil)
}

func openDB(cfg config) (*sql.DB, error) {

	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	//parse the max idle time
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
