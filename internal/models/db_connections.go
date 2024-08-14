package models

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrRecordNotFound = errors.New("record not found")

type ModelsDBConnections struct {
	Examples    ExampleDBConnection
	Users       UserDBConnection
	Tokens      TokenDBConnection
	Permissions PermissionsDBConnection
}

// init all db connections pools for the models
func NewModelsDBConnections(db *sql.DB) ModelsDBConnections {
	return ModelsDBConnections{
		Examples:    ExampleDBConnection{DB: db},
		Users:       UserDBConnection{DB: db},
		Tokens:      TokenDBConnection{DB: db},
		Permissions: PermissionsDBConnection{DB: db},
	}
}

// returns a sql.DB connection pool.
func OpenDB(dsn string, maxOpenConns int, maxIdleConns int, maxIdleTime string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	duration, err := time.ParseDuration(maxIdleTime)
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
