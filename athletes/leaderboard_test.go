package athletes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type storeMock struct{}

func (storeMock) Close()            {}
func (storeMock) Add(Athlete) error { return nil }
func (storeMock) FindAll() (Athletes, error) {
	return Athletes{
		Athlete{"John", "Doe", "d42ebbc6-5b2b-4ff9-83a6-7df87cc20c17", 1},
		Athlete{"Jonah", "Hubbard", "e058c321-b904-46ac-a7fb-9bf0ffeb518e", 2},
		Athlete{"Felicia", "Perez", "32f637d8-40f9-454e-b7b5-88734865cba2", 3},
		Athlete{"Rae", "Burns", "15c95b2b-e63e-442c-98c4-1be4ac871367", 4},
	}, nil
}

type emptyStoreMock struct{}

func (emptyStoreMock) Close()            {}
func (emptyStoreMock) Add(Athlete) error { return nil }
func (emptyStoreMock) FindAll() (Athletes, error) {
	return Athletes{}, nil
}

var initialLeaderboardRows = []LeaderboardRow{
	{Athlete{"John", "Doe", "d42ebbc6-5b2b-4ff9-83a6-7df87cc20c17", 1}, Timings{}},
	{Athlete{"Jonah", "Hubbard", "e058c321-b904-46ac-a7fb-9bf0ffeb518e", 2}, Timings{}},
	{Athlete{"Felicia", "Perez", "32f637d8-40f9-454e-b7b5-88734865cba2", 3}, Timings{}},
	{Athlete{"Rae", "Burns", "15c95b2b-e63e-442c-98c4-1be4ac871367", 4}, Timings{}},
}

func TestInitLeaderboard(t *testing.T) {
	leaderboard, err := NewLeaderboard(&storeMock{})
	assert.Equal(t, nil, err)
	assert.Implements(t, (*Leaderboard)(nil), leaderboard)

	_, err = NewLeaderboard(&emptyStoreMock{})
	assert.Equal(t, "athletes table is empty", err.Error())
}

func TestToLeaderboardRows(t *testing.T) {
	athletes, _ := (&storeMock{}).FindAll()
	actualLeaderboardRows := toLeaderboardRows(athletes)
	assert.Equal(t, initialLeaderboardRows, actualLeaderboardRows)
}

func TestCurrentState(t *testing.T) {
	leaderboard, _ := NewLeaderboard(&storeMock{})
	actualLeaderboardRows := leaderboard.CurrentState()
	assert.Equal(t, initialLeaderboardRows, actualLeaderboardRows)
}

func TestUpdate(t *testing.T) {
	var john = LeaderboardRow{Athlete{"John", "Doe", "d42ebbc6-5b2b-4ff9-83a6-7df87cc20c17", 1}, Timings{}}
	var jonah = LeaderboardRow{Athlete{"Jonah", "Hubbard", "e058c321-b904-46ac-a7fb-9bf0ffeb518e", 2}, Timings{}}
	var felicia = LeaderboardRow{Athlete{"Felicia", "Perez", "32f637d8-40f9-454e-b7b5-88734865cba2", 3}, Timings{}}
	var rae = LeaderboardRow{Athlete{"Rae", "Burns", "15c95b2b-e63e-442c-98c4-1be4ac871367", 4}, Timings{}}

	john.FinishCorridor = "00:01:10.123"
	var updatedLeaderboardRows = []LeaderboardRow{
		john,
		jonah,
		felicia,
		rae,
	}

	leaderboard, _ := NewLeaderboard(&storeMock{})
	updatedRow, err := leaderboard.FindAndUpdate("d42ebbc6-5b2b-4ff9-83a6-7df87cc20c17", "finish_corridor", "00:01:10.123")
	assert.Equal(t, nil, err)
	assert.Equal(t, john, updatedRow)

	actualLeaderboardRows := leaderboard.CurrentState()
	assert.Equal(t, updatedLeaderboardRows, actualLeaderboardRows)

	_, err = leaderboard.FindAndUpdate("non-existing-chip-id", "finish_corridor", "00:01:10.123")
	assert.Equal(t, AtheleteNotFound{"non-existing-chip-id"}, err)
}

func TestLeaderboardSort(t *testing.T) {
	var row LeaderboardRow
	leaderboard, _ := NewLeaderboard(&storeMock{})
	actualLeaderboardRows := leaderboard.CurrentState()
	assert.Equal(t, initialLeaderboardRows, actualLeaderboardRows)

	var john = LeaderboardRow{Athlete{"John", "Doe", "d42ebbc6-5b2b-4ff9-83a6-7df87cc20c17", 1}, Timings{}}
	var jonah = LeaderboardRow{Athlete{"Jonah", "Hubbard", "e058c321-b904-46ac-a7fb-9bf0ffeb518e", 2}, Timings{}}
	var felicia = LeaderboardRow{Athlete{"Felicia", "Perez", "32f637d8-40f9-454e-b7b5-88734865cba2", 3}, Timings{}}
	var rae = LeaderboardRow{Athlete{"Rae", "Burns", "15c95b2b-e63e-442c-98c4-1be4ac871367", 4}, Timings{}}

	// Update 1
	john.FinishCorridor = "00:01:10.342"
	var updatedLeaderboardRows = []LeaderboardRow{
		john,
		jonah,
		felicia,
		rae,
	}
	row, err := leaderboard.FindAndUpdate(john.ChipID, "finish_corridor", "00:01:10.342")
	assert.Equal(t, nil, err)
	assert.Equal(t, john, row)

	actualLeaderboardRows = leaderboard.CurrentState()
	assert.Equal(t, updatedLeaderboardRows, actualLeaderboardRows)

	// Update 2
	felicia.FinishCorridor = "00:01:12.212"
	updatedLeaderboardRows = []LeaderboardRow{
		john,
		felicia,
		jonah,
		rae,
	}
	row, err = leaderboard.FindAndUpdate(felicia.ChipID, "finish_corridor", "00:01:12.212")
	assert.Equal(t, nil, err)
	assert.Equal(t, felicia, row)

	actualLeaderboardRows = leaderboard.CurrentState()
	assert.Equal(t, updatedLeaderboardRows, actualLeaderboardRows)

	// Update 3
	jonah.FinishCorridor = "00:01:13.01"
	updatedLeaderboardRows = []LeaderboardRow{
		john,
		felicia,
		jonah,
		rae,
	}
	row, err = leaderboard.FindAndUpdate(jonah.ChipID, "finish_corridor", "00:01:13.01")
	assert.Equal(t, nil, err)
	assert.Equal(t, jonah, row)

	actualLeaderboardRows = leaderboard.CurrentState()
	assert.Equal(t, updatedLeaderboardRows, actualLeaderboardRows)

	// Update 4
	felicia.FinishLine = "00:01:20.015"
	updatedLeaderboardRows = []LeaderboardRow{
		felicia,
		john,
		jonah,
		rae,
	}
	row, err = leaderboard.FindAndUpdate(felicia.ChipID, "finish_line", "00:01:20.015")
	assert.Equal(t, nil, err)
	assert.Equal(t, felicia, row)

	actualLeaderboardRows = leaderboard.CurrentState()
	assert.Equal(t, updatedLeaderboardRows, actualLeaderboardRows)

	// Update 5
	jonah.FinishLine = "00:01:22.115"
	updatedLeaderboardRows = []LeaderboardRow{
		felicia,
		jonah,
		john,
		rae,
	}
	row, err = leaderboard.FindAndUpdate(jonah.ChipID, "finish_line", "00:01:22.115")
	assert.Equal(t, nil, err)
	assert.Equal(t, jonah, row)

	actualLeaderboardRows = leaderboard.CurrentState()
	assert.Equal(t, updatedLeaderboardRows, actualLeaderboardRows)

	// Update 6
	john.FinishLine = "00:01:25.337"
	updatedLeaderboardRows = []LeaderboardRow{
		felicia,
		jonah,
		john,
		rae,
	}
	row, err = leaderboard.FindAndUpdate(john.ChipID, "finish_line", "00:01:25.337")
	assert.Equal(t, nil, err)
	assert.Equal(t, john, row)

	actualLeaderboardRows = leaderboard.CurrentState()
	assert.Equal(t, updatedLeaderboardRows, actualLeaderboardRows)
}
