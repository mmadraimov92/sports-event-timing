package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"gitlab.com/mooncascade/event-timing-server/athletes"
	"gitlab.com/mooncascade/event-timing-server/router"
)

var (
	buildTime string
	version   string
)

var (
	port         = flag.String("p", "8080", "Port number")
	dbConnection = flag.String("db", os.Getenv("DBURL"), "Postgres database connection string")
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{TimestampFormat: "2006-01-02T15:04:05.9999Z07:00", FullTimestamp: true})

	flag.Parse()
	if *dbConnection == "" {
		*dbConnection = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}
	logger.Infoln("Build:", version, buildTime)

	athletesService, err := athletes.InitService(logger, *dbConnection)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infoln("Listening on", *port)
	http.ListenAndServe(":"+*port, router.New(logger, athletesService))
}
