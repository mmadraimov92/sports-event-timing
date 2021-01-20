package athletes

import (
	"database/sql"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"gitlab.com/mooncascade/event-timing-server/athletes/migrations"
)

const version = 1

// Migrate migrates the Postgres schema to the current version.
func validateSchema(db *sql.DB) error {
	sourceInstance, err := bindata.WithInstance(bindata.Resource(migrations.AssetNames(), migrations.Asset))
	if err != nil {
		return err
	}
	targetInstance, err := postgres.WithInstance(db, new(postgres.Config))
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("go-bindata", sourceInstance, "postgres", targetInstance)
	if err != nil {
		return err
	}
	err = m.Migrate(version)
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return sourceInstance.Close()
}
