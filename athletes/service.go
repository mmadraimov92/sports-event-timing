package athletes

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gitlab.com/mooncascade/event-timing-server/websocket"
)

// Service contains store and validator for Athletes service
type Service struct {
	validator  *validator.Validate
	leadeboard Leaderboard
	logger     *logrus.Logger
	wsManager  websocket.WSManager
}

// InitService initiates store, leaderboard, WSManager and returns Service
func InitService(logger *logrus.Logger, connectionString string) (*Service, error) {
	store, err := NewStore(connectionString)
	if err != nil {
		return nil, fmt.Errorf("store init failed: %w", err)
	}
	defer store.Close()
	l, err := NewLeaderboard(store)
	if err != nil {
		return nil, fmt.Errorf("leaderboard init failed: %w", err)
	}
	wsManager := websocket.NewWSManager()
	service := &Service{validator.New(), l, logger, wsManager}
	return service, nil
}

// Validate validates t
func (s Service) Validate(t interface{}) error {
	return s.validator.Struct(t)
}
