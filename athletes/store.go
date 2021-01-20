package athletes

import (
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
)

// Athlete struct
type Athlete struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	ChipID      string `json:"-"`
	StartNumber int    `json:"start_number"`
}

// Athletes slice
type Athletes []Athlete

// Store interface
//
// FindAll retrieves all Athlete objects from 'athlete' table from db
//
// Add creates new athlete object in db. Used only in testing
//
// Close closes db connection
type Store interface {
	FindAll() (Athletes, error)
	Add(Athlete) error
	Close()
}

// store implements Store
type store struct {
	db *sql.DB
}

func (s store) Close() {
	s.db.Close()
}

const findAllQuery = `
SELECT
	first_name,
	last_name,
	chip_id,
	start_number
FROM athletes
ORDER BY start_number
`

func (s store) FindAll() (Athletes, error) {
	aSlice := Athletes{}
	rows, err := s.db.Query(findAllQuery)
	if err != nil {
		return aSlice, err
	}
	defer rows.Close()
	for rows.Next() {
		a := Athlete{}
		err := rows.Scan(
			&a.FirstName,
			&a.LastName,
			&a.ChipID,
			&a.StartNumber,
		)
		if err != nil {
			return aSlice, err
		}
		aSlice = append(aSlice, a)
	}
	if err := rows.Err(); err != nil {
		return aSlice, err
	}

	return aSlice, err
}

const insertAthleteQuery = `
INSERT INTO athletes (first_name, last_name, start_number, chip_id)
VALUES ($1, $2, $3, $4);
`

func (s store) Add(a Athlete) error {
	_, err := s.db.Exec(insertAthleteQuery, a.FirstName, a.LastName, a.StartNumber, a.ChipID)
	if err != nil {
		return err
	}
	return nil
}

// NewStore initializes db connection, runs migrations then returns Store object
func NewStore(connectionString string) (Store, error) {
	c, err := pgx.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("parsing postgres URI: %w", err)
	}
	db := stdlib.OpenDB(*c)

	if err = validateSchema(db); err != nil {
		return nil, fmt.Errorf("db schema validation: %w", err)
	}

	return &store{db}, nil
}
