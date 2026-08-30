package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coalaura/schgo"
	_ "github.com/mattn/go-sqlite3"
)

const DatabasePath = "data/awesome.db"

type Database struct {
	*sql.DB
}

type StoredCommit struct {
	SHA       string
	Author    string
	Message   string
	AddedURLs []MarkdownURL
	CreatedAt time.Time
}

func LoadDatabase() (*Database, error) {
	dir := filepath.Dir(DatabasePath)

	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

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
	table.Column("author", "TEXT").NotNull()
	table.Column("message", "TEXT").NotNull()
	table.Column("added_urls", "BLOB").NotNull()
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

func (d *Database) GetLatestCommitSha(typ string) (string, error) {
	var sha string

	err := d.QueryRow("SELECT sha FROM commits WHERE type = ? ORDER BY created_at DESC LIMIT 1", typ).Scan(&sha)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}

		return "", err
	}

	return sha, nil
}

func (d *Database) AddNewCommit(typ, sha, author, message string, added []MarkdownURL, createdAt, loadedAt time.Time) error {
	var buffer bytes.Buffer

	err := json.NewEncoder(&buffer).Encode(added)
	if err != nil {
		return err
	}

	_, err = d.Exec(
		"INSERT INTO commits (type, sha, author, message, added_urls, created_at, loaded_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		typ,
		sha,
		author,
		message,
		buffer.Bytes(),
		createdAt.Unix(),
		loadedAt.Unix(),
	)

	return err
}

func (d *Database) GetCommitsByType(ctx context.Context, typ string, filter FeedFilter) ([]StoredCommit, error) {
	addedURLs, arguments := buildAddedURLsExpression(filter)

	query := "SELECT sha, author, message, " + addedURLs + ", created_at FROM commits WHERE type = ? ORDER BY created_at DESC LIMIT 100"
	arguments = append(arguments, typ)

	rows, err := d.QueryContext(ctx, query, arguments...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	defer rows.Close()

	var results []StoredCommit

	for rows.Next() {
		var (
			commit    StoredCommit
			added     []byte
			createdAt int64
		)

		err = rows.Scan(&commit.SHA, &commit.Author, &commit.Message, &added, &createdAt)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(added, &commit.AddedURLs)
		if err != nil {
			return nil, err
		}

		commit.CreatedAt = time.Unix(createdAt, 0)

		results = append(results, commit)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return results, nil
}

func buildAddedURLsExpression(filter FeedFilter) (string, []any) {
	predicates := make([]string, 0)
	arguments := make([]any, 0)

	ownerFilter := filter.RepositoryOwner
	owners := ownerFilter.Include
	operator := "IN"

	if len(ownerFilter.Exclude) > 0 {
		owners = ownerFilter.Exclude
		operator = "NOT IN"
	}

	if len(owners) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(owners)), ",")
		predicate := "substr(repository_path, 1, instr(repository_path, '/') - 1) COLLATE NOCASE " + operator + " (" + placeholders + ")"
		predicates = append(predicates, predicate)

		for _, owner := range owners {
			arguments = append(arguments, owner)
		}
	}

	if len(predicates) == 0 {
		return "added_urls", arguments
	}

	where := strings.Join(predicates, "\n\t\t\t\t\t\tAND ")
	expression := `COALESCE((
					SELECT json_group_array(json(value))
					FROM (
						SELECT value
						FROM (
							SELECT value, key,
								CASE
									WHEN normalized_url LIKE 'https://github.com/%' THEN substr(normalized_url, 20)
									WHEN normalized_url LIKE 'http://github.com/%' THEN substr(normalized_url, 19)
								END AS repository_path
							FROM (
								SELECT value, key, lower(rtrim(json_extract(value, '$.url'), '/')) AS normalized_url
								FROM json_each(commits.added_urls)
							)
						)
						WHERE repository_path GLOB '?*/*?'
							AND repository_path NOT GLOB '*/*/*'
							AND ` + where + `
						ORDER BY key
					)
				), json('[]'))`

	return expression, arguments
}
