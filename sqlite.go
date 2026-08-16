package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/coalaura/schgo"
	_ "github.com/mattn/go-sqlite3"
)

const DatabasePath = "awesome.db"

type Database struct {
	*sql.DB
}

func LoadDatabase() (*Database, error) {
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000&_sync=NORMAL", DatabasePath)

	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(16)
	conn.SetMaxIdleConns(16)
	conn.SetConnMaxLifetime(time.Hour)

	schema, err := schgo.NewSchema(conn)
	if err != nil {
		conn.Close()

		return nil, err
	}

	table := schema.Table("commits")

	table.Primary("id", "INTEGER").AutoIncrement()

	table.Column("type", "TEXT").NotNull()
	table.Column("sha", "TEXT").NotNull()
	table.Column("added_repo", "TEXT")
	table.Column("created_at", "INTEGER").NotNull()
	table.Column("loaded_at", "INTEGER").NotNull()

	table.Index("idx_statistics_type", "type")

	err = schema.Apply()
	if err != nil {
		conn.Close()

		return nil, err
	}

	return &Database{conn}, nil
}
