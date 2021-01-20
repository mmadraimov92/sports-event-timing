package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gitlab.com/mooncascade/event-timing-server/athletes"
	"gitlab.com/mooncascade/event-timing-server/router"
)

var dbConnectionString = ""

func TestMain(m *testing.M) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	resource, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "13.1-alpine",
			Env:        []string{"POSTGRES_USER=postgres", "POSTGRES_PASSWORD=postgres"},
			Mounts:     []string{fmt.Sprintf("%s/init-db.sh:/docker-entrypoint-initdb.d/init-db.sh", dir)},
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
	dbConnectionString = fmt.Sprintf("postgres://postgres:postgres@localhost:%s/postgres?sslmode=disable", resource.GetPort("5432/tcp"))
	if err := pool.Retry(func() error {
		var err error
		c, err := pgx.ParseConfig(dbConnectionString)
		if err != nil {
			return fmt.Errorf("parsing postgres URI: %w", err)
		}
		db := stdlib.OpenDB(*c)
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	os.Exit(code)
}

func TestIntegration(t *testing.T) {
	logger := logrus.New()
	athletesService, err := athletes.InitService(logger, dbConnectionString)
	assert.Equal(t, nil, err)
	r := router.New(logger, athletesService)
	ts := httptest.NewServer(r)
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err.Error())
	}
	u.Scheme = "ws"
	wsURL := fmt.Sprintf("%s/ws", u.String())
	defer ts.Close()

	var john = athletes.LeaderboardRow{Athlete: athletes.Athlete{FirstName: "John", LastName: "Doe", ChipID: "d42ebbc6-5b2b-4ff9-83a6-7df87cc20c17", StartNumber: 1}, Timings: athletes.Timings{}}
	var jonah = athletes.LeaderboardRow{Athlete: athletes.Athlete{FirstName: "Jonah", LastName: "Hubbard", ChipID: "e058c321-b904-46ac-a7fb-9bf0ffeb518e", StartNumber: 2}, Timings: athletes.Timings{}}
	var felicia = athletes.LeaderboardRow{Athlete: athletes.Athlete{FirstName: "Felicia", LastName: "Perez", ChipID: "32f637d8-40f9-454e-b7b5-88734865cba2", StartNumber: 3}, Timings: athletes.Timings{}}
	var rae = athletes.LeaderboardRow{Athlete: athletes.Athlete{FirstName: "Rae", LastName: "Burns", ChipID: "15c95b2b-e63e-442c-98c4-1be4ac871367", StartNumber: 4}, Timings: athletes.Timings{}}
	var leaderboardRows = []athletes.LeaderboardRow{
		john,
		jonah,
		felicia,
		rae,
	}
	// Get initial leaderboard
	resp, body := testRequest(t, ts, "GET", "/leaderboard", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, string(toJSON(t, leaderboardRows)), body)

	// Connect client 1 ws
	client1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("%v: url: %s", err.Error(), wsURL)
	}
	defer client1.Close()

	// Check first ws message for client1
	_, msg, err := client1.ReadMessage()
	assert.Equal(t, nil, err)
	assert.Equal(t, string(toJSON(t, leaderboardRows)), string(msg))

	// Send update 1
	var updatePayload = `
	{
		"chip_id":"d42ebbc6-5b2b-4ff9-83a6-7df87cc20c17",
		"timing_point_id":"finish_corridor",
		"clock_time": "00:01:12.321"
	}
	`
	john.FinishCorridor = "00:01:12.321"
	leaderboardRows = []athletes.LeaderboardRow{
		john,
		jonah,
		felicia,
		rae,
	}
	resp, body = testRequest(t, ts, "POST", "/update", strings.NewReader(updatePayload))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, string(toJSON(t, athletes.SuccessResponse{Message: "updated"})), body)

	// Receive ws update message for client1
	_, msg, err = client1.ReadMessage()
	assert.Equal(t, nil, err)
	assert.Equal(t, string(toJSON(t, john)), string(msg))

	// Connect second ws client
	client2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("%v: url: %s", err.Error(), wsURL)
	}
	defer client2.Close()

	// Check first ws message for client2
	_, msg, err = client2.ReadMessage()
	assert.Equal(t, nil, err)
	assert.Equal(t, string(toJSON(t, leaderboardRows)), string(msg))

	// Send update 2
	updatePayload = `
	{
		"chip_id":"15c95b2b-e63e-442c-98c4-1be4ac871367",
		"timing_point_id":"finish_corridor",
		"clock_time": "00:01:22.321"
	}
	`
	rae.FinishCorridor = "00:01:22.321"
	leaderboardRows = []athletes.LeaderboardRow{
		john,
		rae,
		jonah,
		felicia,
	}
	resp, body = testRequest(t, ts, "POST", "/update", strings.NewReader(updatePayload))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, string(toJSON(t, athletes.SuccessResponse{Message: "updated"})), body)

	// Receive update client1 message
	_, msg, err = client1.ReadMessage()
	assert.Equal(t, nil, err)
	assert.Equal(t, string(toJSON(t, rae)), string(msg))

	// Receive update client2 message
	_, msg, err = client2.ReadMessage()
	assert.Equal(t, nil, err)
	assert.Equal(t, string(toJSON(t, rae)), string(msg))

	// Send update 3
	updatePayload = `
	{
		"chip_id":"32f637d8-40f9-454e-b7b5-88734865cba2",
		"timing_point_id":"finish_corridor",
		"clock_time": "00:01:23"
	}
	`
	felicia.FinishCorridor = "00:01:23"
	leaderboardRows = []athletes.LeaderboardRow{
		john,
		rae,
		felicia,
		jonah,
	}
	resp, body = testRequest(t, ts, "POST", "/update", strings.NewReader(updatePayload))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, string(toJSON(t, athletes.SuccessResponse{Message: "updated"})), body)

	// Send update 4
	updatePayload = `
	{
		"chip_id":"15c95b2b-e63e-442c-98c4-1be4ac871367",
		"timing_point_id":"finish_line",
		"clock_time": "00:01:33"
	}
	`
	rae.FinishLine = "00:01:33"
	leaderboardRows = []athletes.LeaderboardRow{
		rae,
		john,
		felicia,
		jonah,
	}
	resp, body = testRequest(t, ts, "POST", "/update", strings.NewReader(updatePayload))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, string(toJSON(t, athletes.SuccessResponse{Message: "updated"})), body)

	// Ensure correct leaderboard order
	resp, body = testRequest(t, ts, "GET", "/leaderboard", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, string(toJSON(t, leaderboardRows)), body)

	// Missing athlete update
	updatePayload = `
	{
		"chip_id":"aaaaaaaa-e63e-442c-98c4-1be4ac871367",
		"timing_point_id":"finish_corridor",
		"clock_time": "00:01:22.321"
	}
	`
	resp, body = testRequest(t, ts, "POST", "/update", strings.NewReader(updatePayload))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, string(toJSON(t, athletes.ErrorResponse{Error: "athlete with chipId: aaaaaaaa-e63e-442c-98c4-1be4ac871367 not found"})), body)

	// Incomplete update
	var incompleteUpdatePayloads = []string{
		`
		{
			"chip_id":"aaaaaaaa-e63e-442c-98c4-1be4ac871367",
			"timing_point_id":"finish_corridor"
		}
		`,
		`
		{
			"chip_id":"aaaaaaaa-e63e-442c-98c4-1be4ac871367",
			"timing_point_id":"finish_corridor",
			"clock_time": "00"
		}
		`,
		`
		{
			"chip_id":"aaaaaaaa-e63e-442c-98c4-1be4ac871367",
			"timing_point_id":"non-existing-timing-point-id",
			"clock_time": "00:01:22.321"
		}
		`,
		`
		{
			"chip_id":"aaaaaaaa",
			"timing_point_id":"finish_corridor",
			"clock_time": "00:01:22.321"
		}
		`,
	}
	for _, payload := range incompleteUpdatePayloads {
		resp, body = testRequest(t, ts, "POST", "/update", strings.NewReader(payload))
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}

func toJSON(t *testing.T, v interface{}) []byte {
	jsonData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
		return nil
	}
	return jsonData
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
}
