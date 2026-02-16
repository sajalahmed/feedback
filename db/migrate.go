package db

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func newMigrateInstance(dsn string) (*migrate.Migrate, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"mysql",
		driver,
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func RunMigrations(dsn string) error {
	m, err := newMigrateInstance(dsn)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Migrations ran successfully")
	return nil
}

func RevertMigrations(dsn string) error {
	m, err := newMigrateInstance(dsn)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Migrations reverted successfully")
	return nil
}

func ForceVersion(dsn string, version int) error {
	m, err := newMigrateInstance(dsn)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Force(version); err != nil {
		return err
	}

	log.Printf("Forced migration version to %d", version)
	return nil
}
