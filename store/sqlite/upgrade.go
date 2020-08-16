package sqlite

// PostgreSQL specific prefixes, sql
// templates, functions and other helpers

import (
	"context"
	"fmt"
	"github.com/cortezaproject/corteza-server/store/rdbms"
	"github.com/cortezaproject/corteza-server/store/rdbms/ddl"
	"go.uber.org/zap"
)

type (
	upgrader struct {
		s   *Store
		log *zap.Logger
		ddl *ddl.Generator
	}
)

// NewUpgrader returns SQLite schema upgrader
func NewUpgrader(log *zap.Logger, store *Store) *upgrader {
	var g = &upgrader{store, log, ddl.NewGenerator(log)}
	// All modifications we need for the DDL generator
	// to properly support SQLite dialect

	// Cover mysql exceptions
	g.ddl.AddTemplateFunc("columnType", func(ct *ddl.ColumnType) string {
		switch ct.Type {
		case ddl.ColumnTypeTimestamp:
			return "TIMESTAMP"
		default:
			return ddl.GenColumnType(ct)
		}
	})

	return g
}

// Before runs before all tables are upgraded
func (u upgrader) Before(ctx context.Context) error {
	return rdbms.GenericUpgrades(u.log, u).Before(ctx)
}

// After runs after all tables are upgraded
func (u upgrader) After(ctx context.Context) error {
	return rdbms.GenericUpgrades(u.log, u).After(ctx)
}

// CreateTable is triggered for every table defined in the rdbms package
//
// It checks if table is missing and creates it, otherwise
// it runs
func (u upgrader) CreateTable(ctx context.Context, t *ddl.Table) (err error) {
	var exists bool
	if exists, err = u.TableExists(ctx, t.Name); err != nil {
		return
	}

	if !exists {
		if err = u.Exec(ctx, u.ddl.CreateTable(t)); err != nil {
			println(u.ddl.CreateTable(t))
			return err
		}

		for _, i := range t.Indexes {
			if err = u.Exec(ctx, u.ddl.CreateIndex(i)); err != nil {
				return fmt.Errorf("could not create index %s on table %s: %w", i.Name, i.Table, err)
			}
		}
	} else {
		return u.upgradeTable(ctx, t)
	}

	return nil
}

func (u upgrader) Exec(ctx context.Context, sql string, aa ...interface{}) error {
	_, err := u.s.DB().ExecContext(ctx, sql, aa...)
	return err
}

// upgradeTable applies any necessary changes connected to that specific table
func (u upgrader) upgradeTable(ctx context.Context, t *ddl.Table) error {
	g := rdbms.GenericUpgrades(u.log, u)

	switch t.Name {
	default:
		return g.Upgrade(ctx, t)
	}
}

func (u upgrader) TableExists(ctx context.Context, table string) (bool, error) {
	var exists bool

	if err := u.s.DB().GetContext(ctx, &exists, "SELECT COUNT(*) > 0 FROM sqlite_master WHERE type = 'table' AND name = ?", table); err != nil {
		return false, fmt.Errorf("could not check if table exists: %w", err)
	}

	return exists, nil
}

func (u upgrader) AddColumn(ctx context.Context, table string, col *ddl.Column) (added bool, err error) {
	return false, fmt.Errorf("adding column on sqlite tables is not implemented")
}

func (u upgrader) AddPrimaryKey(ctx context.Context, table string, ind *ddl.Index) (added bool, err error) {
	return false, fmt.Errorf("adding primary keys on sqlite tables is not implemented")
}