package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type (
	Config struct {
		Host         string
		Port         int
		Db           string
		User         string
		Password     string
		Schemas      []string
		TLS          bool
		Verbose      bool
		WaitDuration time.Duration
	}

	Connection struct {
		cfg          *Config
		waitDuration time.Duration
		db           *sqlx.DB
	}
)

func New(cfg *Config, retries int) (*Connection, error) {
	var err error
	var db *sqlx.DB
	for retries > 0 {
		db, err = sqlx.Connect("postgres", cfg.address())
		if err != nil {
			retries--
			time.Sleep(cfg.WaitDuration)
			fmt.Fprintf(os.Stderr, "Failed to connect to db (retrying in %v): %v", cfg.WaitDuration, err)
			continue
		}
		break
	}
	if err != nil {
		return nil, err
	}
	return &Connection{
		cfg:          cfg,
		db:           db,
		waitDuration: cfg.WaitDuration,
	}, nil
}

func (c *Connection) HealthCheck() error {
	return c.db.Ping()
}

func (c *Connection) Close() {
	err := c.db.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erorr closing connection: %v", err)
	}
}

func (c *Connection) ChangeOwnerForDB(newOwner, db string) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		c.waitDuration*time.Second)
	defer cancel()
	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	//nolint
	defer tx.Rollback()
	// 1. Alter database
	err = c.changeDBOwner(ctx, tx, newOwner, db)
	if err != nil {
		return err
	}
	// 2. Alter schemas
	err = c.changeSchemasOwners(ctx, tx, newOwner)
	if err != nil {
		return err
	}
	// 2. Alter tables, sequences and other objects inside of schemas
	err = c.changeAllObjectsOwners(ctx, tx, newOwner)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (config Config) address() string {
	address := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=", config.User, config.Password, config.Host, config.Port, config.Db)

	if !config.TLS {
		address += "disable"
	}

	return address
}

func (c *Connection) verboseOutput(statement string) {
	if c.cfg.Verbose {
		fmt.Println(statement)
	}
}
