package athletes

import "fmt"

// AtheleteNotFound .
type AtheleteNotFound struct {
	ChipID string
}

func (a AtheleteNotFound) Error() string {
	return fmt.Sprintf("athlete with chipId: %s not found", a.ChipID)
}
