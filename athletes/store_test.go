package athletes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	var athletesSeed = Athletes{
		Athlete{"John", "Doe", "d42ebbc6-5b2b-4ff9-83a6-7df87cc20c17", 1},
		Athlete{"Jonah", "Hubbard", "e058c321-b904-46ac-a7fb-9bf0ffeb518e", 2},
		Athlete{"Felicia", "Perez", "32f637d8-40f9-454e-b7b5-88734865cba2", 3},
	}

	store, err := NewStore(dbConnectionString)
	assert.Equal(t, nil, err)
	assert.Implements(t, (*Store)(nil), store)

	for _, a := range athletesSeed {
		err = store.Add(a)
		assert.Equal(t, nil, err)
	}

	athletes, err := store.FindAll()
	assert.Equal(t, nil, err)
	assert.Equal(t, athletesSeed, athletes)
	store.Close()

	// Empty DB
	store2, err := NewStore(emptyDBConnectionString)
	assert.Equal(t, nil, err)
	assert.Implements(t, (*Store)(nil), store2)

	athletes, err = store2.FindAll()
	assert.Equal(t, nil, err)
	assert.Equal(t, Athletes{}, athletes)
	store2.Close()
}
