package athletes

import (
	"fmt"
	"sort"
	"time"
)

// Leaderboard interface
//
// CurrentState returns sorted []LeaderboardRow.
//
// FindAndUpdate finds LeaderboardRow by chipID and modifies it.
// Returns modified LeaderboardRow
type Leaderboard interface {
	CurrentState() []LeaderboardRow
	FindAndUpdate(chipID, timingPointID, clockTime string) (LeaderboardRow, error)
}

// LeaderboardRow represents one row on Leaderboard
type LeaderboardRow struct {
	Athlete
	Timings `json:"timings"`
}

// Timings struct contains athlete time for
// finish_corridor and finish_line in 15:04:05.999 format
type Timings struct {
	FinishCorridor string `json:"finish_corridor"`
	FinishLine     string `json:"finish_line"`
}

// leaderboard implements Leaderboard
type leaderboard struct {
	Rows []LeaderboardRow
}

// CurrentState returns current sorted leaderboard
func (l leaderboard) CurrentState() []LeaderboardRow {
	return l.Rows
}

// FindAndUpdate implements Leaderboard.FindAndUpdate
//
// Will return an error if athlete with given chipID was not found
//
// After successful update, l.sort() is called which sorts leaderboard rows by time
// the earliest athlete being first
func (l *leaderboard) FindAndUpdate(chipID, timingPointID, clockTime string) (LeaderboardRow, error) {
	for i, r := range l.Rows {
		if r.ChipID == chipID {
			if timingPointID == "finish_line" {
				l.Rows[i].FinishLine = clockTime
			} else {
				l.Rows[i].FinishCorridor = clockTime
			}
			updatedRow := l.Rows[i]
			l.sort()
			return updatedRow, nil
		}
	}

	return LeaderboardRow{}, AtheleteNotFound{chipID}
}

// sort by LeaderboardRow.FinishLine, LeaderboardRow.FinishCorridor and LeaderboardRow.StartNumber
func (l leaderboard) sort() {
	rows := l.Rows
	sort.Slice(rows, func(i, j int) bool {
		iRow := rows[i]
		jRow := rows[j]
		// Sort by finish_line
		if len(iRow.FinishLine) > 0 || len(jRow.FinishLine) > 0 {
			if len(iRow.FinishLine) == 0 {
				return false
			} else if len(jRow.FinishLine) == 0 {
				return true
			} else {
				iTime, _ := time.Parse("15:04:05.999", iRow.FinishLine)
				jTime, _ := time.Parse("15:04:05.999", jRow.FinishLine)
				if !iTime.Equal(jTime) {
					return iTime.Before(jTime)
				}
			}
		}
		// Sort by finish_corridor
		if len(iRow.FinishCorridor) > 0 || len(jRow.FinishCorridor) > 0 {
			if len(iRow.FinishCorridor) == 0 {
				return false
			} else if len(jRow.FinishCorridor) == 0 {
				return true
			} else {
				iTime, _ := time.Parse("15:04:05.999", iRow.FinishCorridor)
				jTime, _ := time.Parse("15:04:05.999", jRow.FinishCorridor)
				if !iTime.Equal(jTime) {
					return iTime.Before(jTime)
				}
			}
		}
		// Sort by start number
		return rows[i].StartNumber < rows[j].StartNumber
	})
}

// toLeaderboardRows constructs LeaderboardRows from Athletes
func toLeaderboardRows(s Athletes) []LeaderboardRow {
	l := []LeaderboardRow{}
	for _, a := range s {
		l = append(l, LeaderboardRow{a, Timings{}})
	}
	return l
}

// NewLeaderboard initializes Leaderboard object by reading athletes data from store
func NewLeaderboard(s Store) (Leaderboard, error) {
	athletes, err := s.FindAll()
	if err != nil {
		return nil, err
	}
	if len(athletes) == 0 {
		return nil, fmt.Errorf("athletes table is empty")
	}
	return &leaderboard{toLeaderboardRows(athletes)}, nil
}
