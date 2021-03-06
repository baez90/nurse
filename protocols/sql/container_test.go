package sql_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/maxatome/go-testdeep/td"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/baez90/nurse/config"
	"github.com/baez90/nurse/internal/values"
)

const (
	dbName     = "nurse"
	dbUser     = "nurse"
	dbPassword = "asi1EeYi"
)

func PreparePostgresContainer(tb testing.TB) (name string, cfg *config.Server) {
	tb.Helper()

	const postgresPort = "5432/tcp"

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	tb.Cleanup(cancel)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "docker.io/postgres:alpine",
			ExposedPorts: []string{postgresPort},
			SkipReaper:   true,
			AutoRemove:   true,
			Env: map[string]string{
				"POSTGRES_USER":     dbUser,
				"POSTGRES_PASSWORD": dbPassword,
				"POSTGRES_DB":       dbName,
			},
			WaitingFor: wait.ForSQL(postgresPort, "pgx", func(port nat.Port) string {
				return fmt.Sprintf("postgres://%s:%s@localhost:%d/%s", dbUser, dbPassword, port.Int(), dbName)
			}),
		},
		Started: true,
		Logger:  testcontainers.TestLogger(tb),
	})

	td.CmpNoError(tb, err, "testcontainers.GenericContainer()")

	tb.Cleanup(func() {
		td.CmpNoError(tb, container.Terminate(context.Background()), "container.Terminate()")
	})

	ep, err := container.PortEndpoint(ctx, postgresPort, "postgres")
	td.CmpNoError(tb, err, "container.PortEndpoint()")

	srv := new(config.Server)
	td.CmpNoError(tb, srv.UnmarshalURL(ep), "srv.UnmarshalURL()")

	srv.Path = append(srv.Path, dbName)
	srv.Credentials = &config.Credentials{
		Username: dbUser,
		Password: values.StringP(dbPassword),
	}

	name, err = container.Name(ctx)
	td.CmpNoError(tb, err, "container.Name()")

	return name, srv
}

func PrepareMariaDBContainer(tb testing.TB) (name string, cfg *config.Server) {
	tb.Helper()

	const mysqlPort = "3306/tcp"

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	tb.Cleanup(cancel)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "docker.io/mariadb:10",
			ExposedPorts: []string{mysqlPort},
			SkipReaper:   true,
			AutoRemove:   true,
			Env: map[string]string{
				"MARIADB_USER":                 dbUser,
				"MARIADB_PASSWORD":             dbPassword,
				"MARIADB_RANDOM_ROOT_PASSWORD": "1",
				"MARIADB_DATABASE":             dbName,
			},
			WaitingFor: wait.ForSQL(mysqlPort, "mysql", func(port nat.Port) string {
				return fmt.Sprintf("%s:%s@tcp(localhost:%d)/%s", dbUser, dbPassword, port.Int(), dbName)
			}),
		},
		Started: true,
		Logger:  testcontainers.TestLogger(tb),
	})

	td.CmpNoError(tb, err, "testcontainers.GenericContainer()")

	tb.Cleanup(func() {
		td.CmpNoError(tb, container.Terminate(context.Background()), "container.Terminate()")
	})

	ep, err := container.PortEndpoint(ctx, mysqlPort, "mysql")
	td.CmpNoError(tb, err, "container.PortEndpoint()")

	srv := new(config.Server)
	td.CmpNoError(tb, srv.UnmarshalURL(ep), "srv.UnmarshalURL()")

	srv.Path = append(srv.Path, dbName)
	srv.Credentials = &config.Credentials{
		Username: dbUser,
		Password: values.StringP(dbPassword),
	}

	name, err = container.Name(ctx)
	td.CmpNoError(tb, err, "container.Name()")

	return name, srv
}
