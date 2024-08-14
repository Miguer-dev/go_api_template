package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		requestsPerSecond float64
		bucket            int
		enabled           bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		setup     string
		whiteList []string
	}
}

// init config, extract variables from command options (default values on .env variables)
func initConfig() (*config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		return nil, err
	}

	maxOpenConns, err := strconv.Atoi(os.Getenv("DB_MAXOPENCONNS"))
	if err != nil {
		return nil, err
	}

	maxIdleConns, err := strconv.Atoi(os.Getenv("DB_MAXIDLECONNS"))
	if err != nil {
		return nil, err
	}

	requestsPerSecond, err := strconv.ParseFloat(os.Getenv("LIMITER_RPS"), 64)
	if err != nil {
		return nil, err
	}

	bucket, err := strconv.Atoi(os.Getenv("LIMITER_BUCKET"))
	if err != nil {
		return nil, err
	}

	enabled, err := strconv.ParseBool(os.Getenv("LIMITER_ENABLED"))
	if err != nil {
		return nil, err
	}

	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return nil, err
	}

	var cfg config

	flag.IntVar(&cfg.port, "port", port, "API server port")
	flag.StringVar(&cfg.env, "env", os.Getenv("ENV"), "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", maxOpenConns, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", maxIdleConns, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", os.Getenv("DB_MAXIDLETIME"), "PostgreSQL max connection idle time")

	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", enabled, "Enable rate limiter")
	flag.Float64Var(&cfg.limiter.requestsPerSecond, "limiter-rps", requestsPerSecond, "Rate limiter requests per second regeneration")
	flag.IntVar(&cfg.limiter.bucket, "limiter-bucket", bucket, "Rate limiter bucket capacity")

	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", smtpPort, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", os.Getenv("SMTP_SENDER"), "SMTP sender")

	flag.StringVar(&cfg.cors.setup, "cors-setup", os.Getenv("CORS_SETUP"), "CORS policy setup, all = *, specific = origin white list")
	cfg.cors.whiteList = []string{"https://www.example.com", "https://www.example2.com"}
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.whiteList = strings.Fields(val)
		return nil
	})

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	return &cfg, nil
}
