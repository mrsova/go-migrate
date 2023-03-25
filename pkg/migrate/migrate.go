package migrate

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
	"time"
)

type Migration struct {
	id       string
	migrate  func(ctx context.Context, tx pgx.Tx) error
	rollback func(ctx context.Context, tx pgx.Tx) error
}

type Entity struct {
	tablename  string
	migrations []Migration
	pg         *pgxpool.Pool
}

func New(pg *pgxpool.Pool, folder string, tablename string) Entity {
	files, err := os.ReadDir(folder)
	if err != nil {
		log.Fatal(err)
	}
	var migrations []Migration
	for _, f := range files {
		name := strings.Split(f.Name(), ".")
		if name[1] != "sql" {
			continue
		}
		migrate := readMigrationFile(f.Name(), folder+"/"+f.Name())
		if migrate == nil {
			continue
		}
		migrations = append(migrations, *migrate)
	}
	return Entity{
		tablename:  tablename,
		migrations: migrations,
		pg:         pg,
	}
}

func (e *Entity) Migrate(ctx context.Context) error {
	log.Info("creating/checking migrations table...")
	err := e.createMigrationTable(ctx)
	if err != nil {
		return err
	}
	for _, m := range e.migrations {
		var found string
		notEmptyRow := e.pg.QueryRow(ctx, "SELECT id FROM $1 WHERE id=$2", e.tablename, m.id).Scan(&found)
		if notEmptyRow != nil {
			log.Info(fmt.Sprintf("Running migration: %v", m.id))
		} else {
			log.Info(fmt.Sprintf("Skipping migration: %v", m.id))
			continue
		}
		err = e.runMigration(ctx, m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Entity) Rollback(ctx context.Context) error {
	log.Info("creating/checking migrations table...")
	err := e.createMigrationTable(ctx)
	if err != nil {
		return err
	}
	var found string
	notEmptyRow := e.pg.QueryRow(ctx, "SELECT id FROM $1 order by created_at desc limit 1", e.tablename).Scan(&found)
	if notEmptyRow == nil {
		log.Info(fmt.Sprintf("Running rollback migration: %v", found))
	} else {
		log.Info("Skipping rollback migration")
		return nil
	}
	for _, m := range e.migrations {
		if m.id == found {
			errRollback := e.runRollback(ctx, m)
			if errRollback != nil {
				return errRollback
			}
			return nil
		}
		continue
	}
	return nil
}

func (e *Entity) createMigrationTable(ctx context.Context) error {
	_, err := e.pg.Exec(ctx, "CREATE TABLE IF NOT EXISTS $1 (id TEXT PRIMARY KEY, created_at timestamp)", e.tablename)
	if err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}
	return nil
}

func (e *Entity) runMigration(ctx context.Context, m Migration) error {
	tx, err := e.pg.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(ctx, "INSERT INTO $1 (id, created_at) VALUES ($2, $3)", e.tablename, m.id, time.Now())
	if err != nil {
		errRollback := tx.Rollback(ctx)
		log.Fatal(errRollback)
	}
	err = m.migrate(ctx, tx)
	if err != nil {
		errRollback := tx.Rollback(ctx)
		log.Fatal(errRollback)
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (e *Entity) runRollback(ctx context.Context, m Migration) error {
	tx, err := e.pg.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(ctx, "delete from $1 where id = $2", e.tablename, m.id)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	err = m.rollback(ctx, tx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func readMigrationFile(id, file string) *Migration {
	if file == "" {
		return nil
	}
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}
	fileBytes, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	cb := func(filename string, rollback bool) func(ctx context.Context, tx pgx.Tx) error {
		var query string
		s := strings.Split(string(fileBytes), "@DOWN")
		if len(s) != 2 {
			log.Fatalf("error migration %s has not @DOWN", file)
		}
		if !rollback {
			query = strings.TrimSpace(s[0])
		} else {
			query = strings.TrimSpace(s[1])
		}
		if query == "" {
			log.Fatalf("error migration %s for rollback: %t has not sql query", file, rollback)
		}
		return func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, query)
			return err
		}
	}
	return &Migration{
		id:       id,
		migrate:  cb(file, false),
		rollback: cb(file, true),
	}
}
