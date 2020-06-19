package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

const (
	alterDBStatement     = `ALTER DATABASE %s OWNER TO %s`
	alterSchemaStatement = `ALTER SCHEMA %s OWNER TO %s`
	// https://stackoverflow.com/questions/1348126/modify-owner-on-all-tables-simultaneously-in-postgresql/2686185#2686185
	selectTypesStmt = `SELECT 'ALTER TYPE "'|| user_defined_type_schema || '"."' ||
	user_defined_type_name ||'" OWNER TO %s' as stmt FROM information_schema.user_defined_types
	WHERE NOT user_defined_type_schema IN ('pg_catalog', 'information_schema')
	AND user_defined_type_schema IN ('%s')
	ORDER BY user_defined_type_schema, user_defined_type_name`
	selectTablesStmt = `SELECT 'ALTER TABLE "'|| schemaname || '"."' || tablename ||'" OWNER TO %s'
	as stmt FROM pg_tables WHERE NOT schemaname IN
	('pg_catalog', 'information_schema') AND schemaname IN ('%s') ORDER BY schemaname, tablename`
	selectSequenceStmt = `SELECT 'ALTER SEQUENCE "'|| sequence_schema || '"."' || sequence_name ||'" OWNER TO %s'
	as stmt FROM information_schema.sequences WHERE NOT sequence_schema IN ('pg_catalog', 'information_schema')
	AND sequence_schema IN ('%s')
	ORDER BY sequence_schema, sequence_name`
	selectViewsStmt = `SELECT 'ALTER VIEW "'|| table_schema || '"."' || table_name ||'" OWNER TO %s'
	as stmt FROM information_schema.views WHERE NOT table_schema IN ('pg_catalog', 'information_schema')
	AND table_schema IN ('%s')
	ORDER BY table_schema, table_name`
	selectMaterlizedViesStmt = `SELECT 'ALTER TABLE '|| oid::regclass::text ||' OWNER TO %s' as stmt
	FROM pg_class WHERE relkind = 'm' ORDER BY oid`
	selectFunctionsStmt = `SELECT 'ALTER FUNCTION "'|| routine_schema ||'"."' || routine_name ||'" OWNER TO %s' as stmt
	FROM information_schema.routines WHERE routine_body='EXTERNAL'
	AND external_language NOT IN ('INTERNAL','C') AND routine_schema IN ('%s')
	ORDER BY routine_schema, routine_name`
)

// AlterToExec is row structure for any Select Tables, Sequences etc query, which should prepate Alter statement
type AlterToExec struct {
	Stmt string `db:"stmt"`
}

func (c *Connection) changeDBOwner(ctx context.Context, tx *sqlx.Tx, owner, db string) error {
	alter := fmt.Sprintf(alterDBStatement, db, owner)
	c.verboseOutput(alter)
	_, err := tx.ExecContext(ctx, alter)
	if err != nil {
		return err
	}
	return nil
}

func (c *Connection) changeSchemasOwners(ctx context.Context, tx *sqlx.Tx, owner string) error {
	for _, schema := range c.cfg.Schemas {
		alter := fmt.Sprintf(alterSchemaStatement, schema, owner)
		c.verboseOutput(alter)
		_, err := tx.ExecContext(ctx, alter)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Connection) changeAllObjectsOwners(ctx context.Context, tx *sqlx.Tx, owner string) error {
	var (
		// Order from https://www.postgresql.org/docs/current/static/catalog-pg-class.html
		selectStatements = []string{
			fmt.Sprintf(selectTypesStmt, owner, strings.Join(c.cfg.Schemas, "','")),
			fmt.Sprintf(selectTablesStmt, owner, strings.Join(c.cfg.Schemas, "','")),
			fmt.Sprintf(selectSequenceStmt, owner, strings.Join(c.cfg.Schemas, "','")),
			fmt.Sprintf(selectViewsStmt, owner, strings.Join(c.cfg.Schemas, "','")),
			fmt.Sprintf(selectMaterlizedViesStmt, owner),
			fmt.Sprintf(selectFunctionsStmt, owner, strings.Join(c.cfg.Schemas, "','")),
		}
	)
	var err error
	for _, selStmt := range selectStatements {
		err = c.changeObjectsOwner(ctx, tx, selStmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Connection) changeObjectsOwner(ctx context.Context, tx *sqlx.Tx,
	stmt string, args ...interface{}) error {
	finalStmt := fmt.Sprintf(stmt, args...)
	c.verboseOutput(finalStmt)
	stmts := []AlterToExec{}
	err := tx.SelectContext(ctx, &stmts, finalStmt)
	if err != nil {
		return err
	}
	for _, stmt := range stmts {
		c.verboseOutput(stmt.Stmt)
		_, err := tx.ExecContext(ctx, stmt.Stmt)
		if err != nil {
			return err
		}
	}
	return nil
}
