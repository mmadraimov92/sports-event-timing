package athletes

import (
	"encoding/json"
	"errors"
	"net/http"
)

type timingRequest struct {
	ChipID        string `json:"chip_id" validate:"required,uuid4"`
	TimingPointID string `json:"timing_point_id" validate:"required,oneof='finish_corridor' 'finish_line'"`
	ClockTime     string `json:"clock_time" validate:"required,datetime=15:04:05.999"`
}

// ErrorResponse .
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse .
type SuccessResponse struct {
	Message string `json:"message"`
}

// ReceiveTimingEventHandler receives timingRequest, does validation, calls
// Leaderboard.FindAndUpdate, responds with success message and lastly calls
// WSManager.SendMessageToAll notifying all connected ws clients about update
func (s Service) ReceiveTimingEventHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		timingData := timingRequest{}
		err := json.NewDecoder(r.Body).Decode(&timingData)
		if err != nil {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.Validate(timingData); err != nil {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}

		updatedRow, err := s.leadeboard.FindAndUpdate(timingData.ChipID, timingData.TimingPointID, timingData.ClockTime)
		if errors.As(err, &AtheleteNotFound{}) {
			writeError(w, err.Error(), http.StatusNotFound)
			return
		}

		writeSuccess(w, "updated")
		jsonData, err := json.Marshal(updatedRow)
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.wsManager.SendMessageToAll(jsonData)
	}
}

// LeaderboardHandler respons with a sorted array of LeaderboardRows
func (s Service) LeaderboardHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonData, err := json.Marshal(s.leadeboard.CurrentState())
		if err != nil {
			s.logger.Errorln(err.Error())
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, jsonData, http.StatusOK)
	}
}

// WSHandler handles websocket connection, adds new client by calling WSManager.AddCLient,
// sends current leaderboard as first message to client and lastly calls WSManager.StartClient
func (s Service) WSHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := s.wsManager.Upgrade(w, r, nil)
		if err != nil {
			s.logger.Errorln(err.Error())
			return
		}
		clientID := s.wsManager.AddClient(ws, s.logger)
		jsonData, err := json.Marshal(s.leadeboard.CurrentState())
		if err != nil {
			s.logger.Errorln(err.Error())
			return
		}
		go s.wsManager.SendMessageToOne(jsonData, clientID)
		s.wsManager.StartClient(clientID)
	}
}

func writeError(w http.ResponseWriter, err string, code int) {
	jsonData, _ := json.Marshal(ErrorResponse{err})
	writeJSON(w, jsonData, code)
}

func writeSuccess(w http.ResponseWriter, message string) {
	jsonData, _ := json.Marshal(SuccessResponse{message})
	writeJSON(w, jsonData, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, jsonData []byte, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(jsonData)
}
