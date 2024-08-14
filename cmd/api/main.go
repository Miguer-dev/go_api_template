package main

import (
	"log"
	"sync"

	_ "github.com/lib/pq"

	"go.api.template/internal/mailer"
	"go.api.template/internal/models"
	"go.api.template/internal/vcs"
)

var (
	version = vcs.Version()
)

type application struct {
	config   *config
	errorLog *log.Logger
	infoLog  *log.Logger
	models   models.ModelsDBConnections
	mailer   mailer.Mailer
	wg       *sync.WaitGroup
}

func main() {
	errorLog, infoLog := initLogs()

	cfg, err := initConfig()
	if err != nil {
		errorLog.Fatal(err.Error())
	}

	db, err := models.OpenDB(cfg.db.dsn, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		errorLog.Fatal(err.Error())
	}
	defer db.Close()
	infoLog.Printf("database connection pool established")

	initMetrics(db)

	app := &application{
		config:   cfg,
		errorLog: errorLog,
		infoLog:  infoLog,
		models:   models.NewModelsDBConnections(db),
		mailer:   mailer.InitMailer(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		wg:       &sync.WaitGroup{},
	}

	err = app.serve()
	if err != nil {
		errorLog.Fatal(err.Error())
	}
}
