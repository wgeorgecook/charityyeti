package main

import (
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// pgClient is the bun database we use to query
// against postgres
var pgClient *bun.DB

func initPostgres() {
		// Open a PostgreSQL database.
		pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(cfg.PostgresConnectionURI)))
	
		// Create a Bun db on top of it.
		pgClient = bun.NewDB(pgdb, pgdialect.New())
}

func cleanupPostgres() {
	
}