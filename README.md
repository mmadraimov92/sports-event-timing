# sports-event-timing

Test task.

Server that manages an automatic timing system for the finish corridor and finish line.

List of participants with their respective `chipId` is read from database from `athletes` table. Server keeps internal `leaderboard` and serves updates to connected clients via WebSocket.

## API

1. GET `/leaderboard` - get current leaderboard
2. POST `/update` - post an timing event update
3. GET `/ws` - connect to WebSocket to subscribe for updates
4. GET `/openapi` - openapi specs

For more details go to `localhost:8080/openapi` after starting servver

## Quick start

`docker-compose up`. This will start server and postgres db. Default port is `8080`

## Dependencies

`go mod download`

## Tests

`make unit-test` and `make integration-test`

## Server arguments

1. `-p` - port number, default `8080`
2. `-db` - database connection string, if not specified will get value from `DBURL` env variable. Default value `postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable`
