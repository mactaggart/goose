package goose

import (
	"database/sql"
)

// Redo rolls back the most recently applied migration, then runs it again.
func (in *Instance) Redo(db *sql.DB, dir string, opts ...OptionsFunc) error {
	option := &options{}
	for _, f := range opts {
		f(option)
	}
	migrations, err := in.CollectMigrations(dir, minVersion, maxVersion)
	if err != nil {
		return err
	}
	var (
		currentVersion int64
	)
	if option.noVersioning {
		if len(migrations) == 0 {
			return nil
		}
		currentVersion = migrations[len(migrations)-1].Version
	} else {
		if currentVersion, err = in.GetDBVersion(db); err != nil {
			return err
		}
	}

	current, err := migrations.Current(currentVersion)
	if err != nil {
		return err
	}
	current.noVersioning = option.noVersioning

	if err := current.Down(db); err != nil {
		return err
	}
	if err := current.Up(db); err != nil {
		return err
	}
	return nil
}
