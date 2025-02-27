package e2e

import (
	"database/sql"
	"testing"

	"github.com/mactaggart/goose/v3"
	"github.com/matryer/is"
)

func TestNoVersioning(t *testing.T) {
	if *dialect != dialectPostgres {
		t.SkipNow()
	}
	const (
		// Total owners created by the seed files.
		wantSeedOwnerCount = 250
		// These are owners created by migration files.
		wantOwnerCount = 4
	)
	is := is.New(t)
	db, err := newDockerDB(t)
	is.NoErr(err)

	err = in.Up(db, migrationsDir)
	is.NoErr(err)
	baseVersion, err := in.GetDBVersion(db)
	is.NoErr(err)

	t.Run("seed-up-down-to-zero", func(t *testing.T) {
		is := is.NewRelaxed(t)
		// Run (all) up migrations from the seed dir
		{
			err = in.Up(db, seedDir, goose.WithNoVersioning())
			is.NoErr(err)
			// Confirm no changes to the versioned schema in the DB
			currentVersion, err := in.GetDBVersion(db)
			is.NoErr(err)
			is.Equal(baseVersion, currentVersion)
			seedOwnerCount, err := countSeedOwners(db)
			is.NoErr(err)
			is.Equal(seedOwnerCount, wantSeedOwnerCount)
		}

		// Run (all) down migrations from the seed dir
		{
			err = in.DownTo(db, seedDir, 0, goose.WithNoVersioning())
			is.NoErr(err)
			// Confirm no changes to the versioned schema in the DB
			currentVersion, err := in.GetDBVersion(db)
			is.NoErr(err)
			is.Equal(baseVersion, currentVersion)
			seedOwnerCount, err := countSeedOwners(db)
			is.NoErr(err)
			is.Equal(seedOwnerCount, 0)
		}

		// The migrations added 4 non-seed owners, they must remain
		// in the database afterwards
		ownerCount, err := countOwners(db)
		is.NoErr(err)
		is.Equal(ownerCount, wantOwnerCount)
	})

	t.Run("test-seed-up-reset", func(t *testing.T) {
		is := is.NewRelaxed(t)
		// Run (all) up migrations from the seed dir
		{
			err = in.Up(db, seedDir, goose.WithNoVersioning())
			is.NoErr(err)
			// Confirm no changes to the versioned schema in the DB
			currentVersion, err := in.GetDBVersion(db)
			is.NoErr(err)
			is.Equal(baseVersion, currentVersion)
			seedOwnerCount, err := countSeedOwners(db)
			is.NoErr(err)
			is.Equal(seedOwnerCount, wantSeedOwnerCount)
		}

		// Run reset (effectively the same as down-to 0)
		{
			err = in.Reset(db, seedDir, goose.WithNoVersioning())
			is.NoErr(err)
			// Confirm no changes to the versioned schema in the DB
			currentVersion, err := in.GetDBVersion(db)
			is.NoErr(err)
			is.Equal(baseVersion, currentVersion)
			seedOwnerCount, err := countSeedOwners(db)
			is.NoErr(err)
			is.Equal(seedOwnerCount, 0)
		}

		// The migrations added 4 non-seed owners, they must remain
		// in the database afterwards
		ownerCount, err := countOwners(db)
		is.NoErr(err)
		is.Equal(ownerCount, wantOwnerCount)
	})

	t.Run("test-seed-up-redo", func(t *testing.T) {
		is := is.NewRelaxed(t)
		// Run (all) up migrations from the seed dir
		{
			err = in.Up(db, seedDir, goose.WithNoVersioning())
			is.NoErr(err)
			// Confirm no changes to the versioned schema in the DB
			currentVersion, err := in.GetDBVersion(db)
			is.NoErr(err)
			is.Equal(baseVersion, currentVersion)
			seedOwnerCount, err := countSeedOwners(db)
			is.NoErr(err)
			is.Equal(seedOwnerCount, wantSeedOwnerCount)
		}

		// Run reset (effectively the same as down-to 0)
		{
			err = in.Redo(db, seedDir, goose.WithNoVersioning())
			is.NoErr(err)
			// Confirm no changes to the versioned schema in the DB
			currentVersion, err := in.GetDBVersion(db)
			is.NoErr(err)
			is.Equal(baseVersion, currentVersion)
			seedOwnerCount, err := countSeedOwners(db)
			is.NoErr(err)
			is.Equal(seedOwnerCount, wantSeedOwnerCount) // owners should be unchanged
		}

		// The migrations added 4 non-seed owners, they must remain
		// in the database afterwards along with the 250 seed owners for a
		// total of 254.
		ownerCount, err := countOwners(db)
		is.NoErr(err)
		is.Equal(ownerCount, wantOwnerCount+wantSeedOwnerCount)
	})
}

func countSeedOwners(db *sql.DB) (int, error) {
	q := `SELECT count(*)FROM owners WHERE owner_name LIKE'seed-user-%'`
	var count int
	if err := db.QueryRow(q).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func countOwners(db *sql.DB) (int, error) {
	q := `SELECT count(*)FROM owners`
	var count int
	if err := db.QueryRow(q).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
