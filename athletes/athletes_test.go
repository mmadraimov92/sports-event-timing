package athletes

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var dbConnectionString = ""
var emptyDBConnectionString = ""

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	resource1, connString1 := startDB(pool)
	dbConnectionString = connString1
	resource2, connString2 := startDB(pool)
	emptyDBConnectionString = connString2
	code := m.Run()
	if err := pool.Purge(resource1); err != nil {
		log.Fatalf("Could not purge resource1: %s", err)
	}
	if err := pool.Purge(resource2); err != nil {
		log.Fatalf("Could not purge resource2: %s", err)
	}

	os.Exit(code)
}

func startDB(pool *dockertest.Pool) (*dockertest.Resource, string) {
	resource, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "13.1-alpine",
			Env:        []string{"POSTGRES_USER=postgres", "POSTGRES_PASSWORD=postgres"},
		}, func(hc *docker.HostConfig) {
			hc.AutoRemove = true
			hc.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		},
	)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	resource.Expire(120)
	connectionString := fmt.Sprintf("postgres://postgres:postgres@localhost:%s/postgres?sslmode=disable", resource.GetPort("5432/tcp"))
	if err := pool.Retry(func() error {
		var err error
		c, err := pgx.ParseConfig(connectionString)
		if err != nil {
			return fmt.Errorf("parsing postgres URI: %w", err)
		}
		db := stdlib.OpenDB(*c)
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	return resource, connectionString
}
